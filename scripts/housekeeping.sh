#!/usr/bin/env bash
# housekeeping.sh: keep the openpak org under its storage quota. Artifacts: keep those of
# the newest $KEEP runs of each workflow in each repo, and none older than $MAX_AGE_DAYS days
# (the quota filled up and blocked releases on 2026-09-23 even with $KEEP runs kept).
# Container images: keep the newest $KEEP tagged versions of each package plus the platform/attestation manifests their
# indexes reference; delete every other version. Production pulls :latest (always kept)
# or a pinned tag that is the newest release, so $KEEP >= 2 never removes a running image.
#
# DRY_RUN=1 (default) only prints. Needs GH_TOKEN with repo, read:packages, delete:packages.
set -euo pipefail
ORG=${ORG:-openpak}; KEEP=${KEEP:-4}; DRY_RUN=${DRY_RUN:-1}; MAX_AGE_DAYS=${MAX_AGE_DAYS:-3}
cutoff=$(date -u -d "-${MAX_AGE_DAYS} days" +%Y-%m-%dT%H:%M:%SZ)

del() { if [ "$DRY_RUN" = 1 ]; then echo "would delete $1"; else gh api -X DELETE "$1" >/dev/null && echo "deleted $1"; fi; }

# --- artifacts ---
for repo in $(gh repo list "$ORG" --limit 500 --json name -q '.[].name'); do
  gh api "repos/$ORG/$repo/actions/artifacts?per_page=100" --paginate \
    -q '.artifacts[]|select(.expired==false)|"\(.id) \(.workflow_run.id) \(.created_at)"' > /tmp/a.txt || continue
  [ -s /tmp/a.txt ] || continue
  # newest time per run, then the run's workflow, then the newest $KEEP runs per workflow
  keep=$(sort -k2,2 -k3,3r /tmp/a.txt | awk '!s[$2]++{print $2, $3}' | while read -r run t; do
      wf=$(gh api "repos/$ORG/$repo/actions/runs/$run" -q .workflow_id 2>/dev/null || echo gone)
      echo "$wf $t $run"; done | sort -k1,1 -k2,2r | awk -v k="$KEEP" '{n[$1]++} n[$1]<=k{print $3}')
  while read -r id run created; do
    # ISO-8601 UTC timestamps compare correctly as strings
    if [[ "$created" < "$cutoff" ]] || ! grep -qx "$run" <<<"$keep"; then
      del "repos/$ORG/$repo/actions/artifacts/$id"
    fi
  done < /tmp/a.txt
done

# --- container images ---
accept='application/vnd.oci.image.index.v1+json,application/vnd.docker.distribution.manifest.list.v2+json,application/vnd.oci.image.manifest.v1+json,application/vnd.docker.distribution.manifest.v2+json'
# The org package list omits some packages (npln, mods, nn-* were missing on 2026-09-18),
# so every repo name is tried as a package too; a name with no package is skipped.
pkgs=$( { gh api "orgs/$ORG/packages?package_type=container&per_page=100" --paginate -q '.[].name'
          gh repo list "$ORG" --limit 500 --json name -q '.[].name'; } | sort -u)
for pkg in $pkgs; do
  api="orgs/$ORG/packages/container/$pkg/versions"
  gh api "$api?per_page=100" --paginate -q '.[]|"\(.id) \(.name) \(.created_at) \(.metadata.container.tags|length)"' > /tmp/v.txt 2>/dev/null || continue
  kept=$(awk '$4>0' /tmp/v.txt | sort -k3,3r | head -n "$KEEP" | awk '{print $2}')
  [ -n "$kept" ] || { echo "$pkg: no tagged versions, skipped"; continue; }
  tok=$(curl -fsu "x:$GH_TOKEN" "https://ghcr.io/token?scope=repository:$ORG/$pkg:pull" | jq -r .token)
  keepset=$kept; ok=1
  for d in $kept; do
    # an index's children are untagged versions that must stay with it
    m=$(curl -fs -H "Authorization: Bearer $tok" -H "Accept: $accept" "https://ghcr.io/v2/$ORG/$pkg/manifests/$d") || { ok=0; break; }
    keepset+=$'\n'$(jq -r '.manifests[]?.digest' <<<"$m")
  done
  # never delete on a partial picture: a missed child breaks a kept image
  [ "$ok" = 1 ] || { echo "$pkg: could not read a kept manifest, skipped"; continue; }
  while read -r id digest _; do
    grep -qx "$digest" <<<"$keepset" || del "$api/$id"
  done < /tmp/v.txt
done
