#!/usr/bin/env bash
# Sync upstream assigned issues into .scratch/issues/assigned/ (local mirror only).
set -euo pipefail

UPSTREAM_REPO="${UPSTREAM_ISSUES_REPO:-ferro-labs/ai-gateway}"
UPSTREAM_OWNER="${UPSTREAM_REPO%%/*}"
UPSTREAM_NAME="${UPSTREAM_REPO#*/}"
MIRROR_DIR="${ASSIGNED_ISSUES_MIRROR_DIR:-.scratch/issues/assigned}"
MAX_ISSUES="${ASSIGNED_ISSUES_LIMIT:-100}"

slugify() {
  echo "$1" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+|-+$//g' | cut -c1-72
}

issue_number_from_mirror_path() {
  local base="${1##*/}"
  local issue_num="${base%%-*}"
  if [[ "$issue_num" =~ ^[0-9]+$ ]]; then
    echo "$issue_num"
  fi
}

# Returns 0 when any closing PR merged into a release/* branch.
issue_merged_to_release() {
  local number="$1"
  local merged_json
  if ! merged_json="$(gh api graphql -f query="
    query {
      repository(owner:\"${UPSTREAM_OWNER}\", name:\"${UPSTREAM_NAME}\") {
        issue(number:${number}) {
          closedByPullRequestsReferences(first:20) {
            nodes { merged baseRefName }
          }
        }
      }
    }" 2>/dev/null)"; then
    return 1
  fi

  echo "$merged_json" | jq -e '
    .data.repository.issue.closedByPullRequestsReferences.nodes
    | any(.merged == true and (.baseRefName | startswith("release/")))
  ' >/dev/null
}

remove_mirror_files_for_issue() {
  local number="$1"
  find "$MIRROR_DIR" -maxdepth 1 -type f -name "${number}-*.md" -delete 2>/dev/null || true
}

sync_assigned_issues() {
  if ! command -v gh >/dev/null 2>&1; then
    echo "gh CLI not found; skipping assigned-issue sync" >&2
    return 1
  fi
  if ! command -v jq >/dev/null 2>&1; then
    echo "jq not found; skipping assigned-issue sync" >&2
    return 1
  fi

  mkdir -p "$MIRROR_DIR"

  local issues_json
  if ! issues_json="$(gh issue list \
    --repo "$UPSTREAM_REPO" \
    --assignee @me \
    --state all \
    --limit "$MAX_ISSUES" \
    --json number,title,state,updatedAt)"; then
    echo "gh issue list failed; skipping assigned-issue sync" >&2
    return 1
  fi

  local count
  count="$(echo "$issues_json" | jq 'length')"
  if [[ "$count" -eq 0 ]]; then
    find "$MIRROR_DIR" -maxdepth 1 -type f -name '[0-9][0-9]*-*.md' -delete 2>/dev/null || true
    echo "Synced 0 assigned issues from $UPSTREAM_REPO"
    return 0
  fi

  local active_numbers=()
  local removed_lines=()
  local summary_lines=()
  summary_lines+=("Assigned issues mirrored from $UPSTREAM_REPO (source: gh --assignee @me):")

  while IFS= read -r issue; do
    local number title
    number="$(echo "$issue" | jq -r '.number')"
    title="$(echo "$issue" | jq -r '.title')"

    if issue_merged_to_release "$number"; then
      remove_mirror_files_for_issue "$number"
      removed_lines+=("- #${number}: ${title} (merged to release/*; mirror removed)")
      continue
    fi

    local detail
    if ! detail="$(gh issue view "$number" \
      --repo "$UPSTREAM_REPO" \
      --json number,title,state,body,createdAt,updatedAt,labels,assignees,url)"; then
      echo "gh issue view #$number failed; skipping" >&2
      continue
    fi

    local slug path
    slug="$(slugify "$title")"
    path="${MIRROR_DIR}/${number}-${slug}.md"

    local url state created updated labels assignees body
    url="$(echo "$detail" | jq -r '.url')"
    state="$(echo "$detail" | jq -r '.state')"
    created="$(echo "$detail" | jq -r '.createdAt')"
    updated="$(echo "$detail" | jq -r '.updatedAt')"
    labels="$(echo "$detail" | jq -r '[.labels[].name] | join(", ")')"
    assignees="$(echo "$detail" | jq -r '[.assignees[].login] | join(", ")')"
    body="$(echo "$detail" | jq -r '.body // ""')"

    {
      printf '# Issue #%s: %s\n\n' "$number" "$title"
      printf -- '- URL: %s\n' "$url"
      printf -- '- State: %s\n' "$state"
      printf -- '- Created: %s\n' "$created"
      printf -- '- Updated: %s\n' "$updated"
      printf -- '- Labels: %s\n' "$labels"
      printf -- '- Assignee: %s\n\n' "$assignees"
      if [[ -n "$body" ]]; then
        printf '%s\n' "$body"
      fi
    } >"$path"

    # Drop stale slug variants when the title changes.
    find "$MIRROR_DIR" -maxdepth 1 -type f -name "${number}-*.md" ! -path "$path" -delete 2>/dev/null || true

    active_numbers+=("$number")
    summary_lines+=("- #${number}: ${title} ($(echo "$state" | tr '[:upper:]' '[:lower:]')) -> ${path}")
  done < <(echo "$issues_json" | jq -c '.[]')

  local active_pattern=""
  if ((${#active_numbers[@]} > 0)); then
    active_pattern="$(printf '%s\n' "${active_numbers[@]}" | paste -sd'|' -)"
  fi

  while IFS= read -r existing; do
    [[ -n "$existing" ]] || continue
    local issue_num
    issue_num="$(issue_number_from_mirror_path "$existing")"
    [[ -n "$issue_num" ]] || continue

    if ((${#active_numbers[@]} > 0)) && [[ "$issue_num" =~ ^(${active_pattern})$ ]]; then
      continue
    fi

    if issue_merged_to_release "$issue_num"; then
      remove_mirror_files_for_issue "$issue_num"
      removed_lines+=("- #${issue_num}: mirror removed (merged to release/*)")
      continue
    fi

    rm -f "$existing"
    removed_lines+=("- #${issue_num}: mirror removed (no longer assigned)")
  done < <(find "$MIRROR_DIR" -maxdepth 1 -type f -name '[0-9][0-9]*-*.md' -print 2>/dev/null || true)

  if ((${#removed_lines[@]} > 0)); then
    summary_lines+=("")
    summary_lines+=("Removed mirrors:")
    summary_lines+=("${removed_lines[@]}")
  fi

  printf '%s\n' "${summary_lines[@]}"
  echo "Synced ${#active_numbers[@]} assigned issue(s) into ${MIRROR_DIR}/"
}

main() {
  local hook_input=""
  if [[ -t 0 ]]; then
    :
  else
    hook_input="$(cat)"
  fi

  local summary=""
  if summary="$(sync_assigned_issues)"; then
    :
  else
    summary="Assigned-issue sync skipped (see Hooks output channel). Mirror path: ${MIRROR_DIR}/"
  fi

  # sessionStart: best-effort context injection + file sync on disk.
  if [[ -n "$hook_input" ]] && echo "$hook_input" | jq -e '.session_id? // empty' >/dev/null 2>&1; then
    local context
    context="$(printf '%s\n\nRead mirrored tickets under %s/ before triaging or /to-issues work.' "$summary" "$MIRROR_DIR")"
    jq -n --arg additional_context "$context" '{additional_context: $additional_context}'
    exit 0
  fi

  # Manual invocation or other callers: print human-readable summary only.
  printf '%s\n' "$summary"
  exit 0
}

main "$@"
