# Title

Ubiquity Issue Triage Process

## Status

provisional

## Summary

In order to ensure that issues reported by Ubiquity users are reviewed on
a consistent basis, we should meet on a regular schedule in a live
meeting to review newly submitted issues, and on some recurring basis
look at potentially stale issues for consideration whether it should be
closed, increase priority, etc.

## Proposal

During the triage process, the moderator should go through each of the
subcategories listed below and apply the process to each issue.

### New Issue Triage

[GitHub Search
Query](https://github.com/ubiquitycluster/ubiquity/issues?utf8=%E2%9C%93&q=archived%3Afalse+no%3Alabel+is%3Aissue+sort%3Acreated-asc+is%3Aopen):
`archived:false no:label is:issue sort:created-asc
is:open`

- Evalulate if the issue is still relevant.
  - If not, close the issue.
- Determine the kind, and apply the right label. For example: bug, feature, etc.
- Make a best guess at priority, if the issue isn't actively being
  worked on
- If needed, ask for more information from the reporter or a
  developer. Label this issue `priority/awaiting-evidence`.
- Mark trivial issues as `good first issue`

### Awaiting Evidence

[GitHub Search
Query](https://github.com/ubiquitycluster/ubiquity/issues?utf8=%E2%9C%93&q=archived%3Afalse+is%3Aissue+sort%3Acreated-asc+is%3Aopen+label%3Apriority%2Fawaiting-more-evidence):`archived:false
 is:issue sort:created-asc is:open
label:priority/awaiting-more-evidence`

- Review if the required evidence has been provided, if so, change the
  priority/kind as needed, or close the issue if resolved.

### Stale Issues

[GitHub Search
Query](https://github.com/ubiquitycluster/ubiquity/issues?q=archived%3Afalse+is%3Aissue+sort%3Acreated-asc+is%3Aopen+label%3Alifecycle%2Fstale):
`archived:false is:issue sort:created-asc is:open
label:lifecycle/stale`