# How to Contribute

Ubiquity project components are [FSL-1.0-Apache-2.0 licensed](LICENSE.md) and accept contributions via
GitHub pull requests.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Certificate of Origin](#certificate-of-origin)
- [Finding Things That Need Help](#finding-things-that-need-help)
- [Versioning](#versioning)
   - [Codebase](#codebase)
      - [Backporting](#backporting)
- [Branches](#branches)
- [Contributing a Patch](#contributing-a-patch)
- [Backporting a Patch](#backporting-a-patch)
- [Breaking Changes](#breaking-changes)
   - [Merge Approval](#merge-approval)
   - [Google Doc Viewing Permissions](#google-doc-viewing-permissions)
   - [Issue and Pull Request Management](#issue-and-pull-request-management)
   - [Commands and Workflow](#commands-and-workflow)
- [Release Process](#release-process)
   - [Exact Steps](#exact-steps)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Contributor Licence Agreement

By contributing to this project you agree to the [Contributor License Agreement](CLA.md).
We want to ensure that the Ubiquity project continues to be a great place to collaborate
and improve the project, with a clear and transparent process for contributing.

Because the project is licensed under Apache 2.0, you must ensure that you have the right
to contribute any code that you submit to the project. Doing so affirms that:

- The contribution was created in whole or in part by you and you have the right to submit it under the open source license indicated in the [LICENSE](LICENSE.md) file; or
- The contribution is based upon previous work that, to the best of your knowledge, is covered under an appropriate open source license and you have the right under that license to submit that work with modifications, whether created in whole or in part by you, under the same open source license (unless you are permitted to submit under a different license), as indicated in the [LICENSE](LICENSE.md) file; or
- The contribution was provided directly to you by some other person who certified (a), (b) or (c) and you have not modified it.

## Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. See the [DCO](DCO) file for details.

## Finding Things That Need Help

If you're new to the project and want to help, but don't know where to start, we
have a semi-curated list of issues that
should not need deep knowledge of the system. [Have a look and see if anything
sounds interesting](https://github.com/logicalisuki/ubiquity/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22).
Alternatively, read some of the docs and try to write your own application helm chart,
file and fix any/all issues that come up, including gaps in documentation!

## Versioning

### Codebase

Ubiquity doesn't follow release cadence and versioning of upstream Kubernetes semantic versioning. This codebase guarantees the following:

- A (*minor*) release CAN include:
   - Introduction of new API versions, or new Kinds.
   - Compatible API changes like field additions, deprecation notices, etc.
   - Breaking API changes for deprecated APIs, fields, or code.
   - Features, promotion or removal of feature gates.
   - And more!

- A (*patch*) release SHOULD only include backwards compatible set of bugfixes.

These guarantees extend to all code exposed in our Go Module, including
*types from dependencies in public APIs*.
Types and functions not in public APIs are not considered part of the guarantee.

#### Backporting

**Note: At present we only have main branch, there is no other release branch. As such backporting is not
an option currently. Once we have release branches the following should be considered.**

We only accept backports of critical bugs, security issues, or bugs without easy workarounds, any
backport MUST not be breaking for either API or behavioral changes.
We generally do not accept PRs against older release branches.

## Branches

Ubiquity has only the *main* branch currently.

The goal is to have two types of branches in future: the *main* branch and *release-X* branches.

The *main* branch is where development happens. All the latest and
greatest code, including breaking changes, happens on main. Currently patch releases are done from this branch
only.

The *release-X* branches will contain stable, backwards compatible code. On every
major or minor release, a new branch will be created. It is from these
branches that minor and patch releases will be tagged. In some cases, it may
be necessary to open PRs for bugfixes directly against stable branches, but
this should generally not be the case.

## Contributing to Ubiquity

All PRs submitted to the Ubiquity project must be signed, see below for more information

## Sign your work

The sign-off is a simple line at the end of the explanation for the patch. Your
signature certifies that you wrote the patch or otherwise have the right to pass
it on as an open-source patch. The rules are pretty simple: if you can certify
the below (from [developercertificate.org](http://developercertificate.org/)):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.

Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

Then you just add a line to every git commit message:

    Signed-off-by: Joe Smith <joe.bloggs@email.com>

Use your real name (sorry, no pseudonyms or anonymous contributions.)

If you set your `user.name` and `user.email` git configs, you can sign your
commit automatically with `git commit -s`.


1. Fork the desired repo, develop and test your code changes.
1. Submit a pull request.

All code PR should be labeled with one of

- ⚠️ (`:warning:`, major or breaking changes)
- ✨ (`:sparkles:`, feature additions)
- 🐛 (`:bug:`, patch and bugfixes)
- 📖 (`:book:`, documentation or proposals)
- 🌱 (`:seedling:`, minor or other)

Individual commits should not be tagged separately, but will generally be
assumed to match the PR. For instance, if you have a bugfix in with
a breaking change, it's generally encouraged to submit the bugfix
separately, but if you must put them in one PR, mark the commit
separately.

All changes must be code reviewed. Coding conventions and standards are
explained in the official [developer
docs](https://github.com/kubernetes/community/tree/master/contributors/devel).
Expect reviewers to request that you
avoid common [go style
mistakes](https://github.com/golang/go/wiki/CodeReviewComments) in your PRs.

## Backporting a Patch

Ubiquity does not yet maintain branches other than `main`. However, in future there will be
release branches as well.

Once we will have release branches in future, Ubiquity will maintain older versions through `release-X.Y` branches. We accept backports of bug fixes to the most recent
release branch. For example, if the most recent branch is `release-0.2`, and the
`main` branch is under active
development for v0.3.0, a bug fix that merged to `main` that also affects
`v0.2.x` may be considered for backporting
to `release-0.2`. We generally do not accept PRs against older release branches.

## Breaking Changes

Breaking changes are generally allowed in the `main` branch.

There may be times, however, when `main` is closed for breaking changes. This
is likely to happen as we are close to release a new minor version.

Breaking changes are not allowed in release branches, as these represent minor
versions that have already been released.
These versions have consumers who expect the APIs, behaviors, etc. to remain
stable during the life time of the patch stream for the minor release.

Examples of breaking changes include:

- Removing or renaming a field in a helm chart values file
- Removing or renaming a helm chart
- Removing or renaming an exported constant, variable, type, or function
- Updating the version of critical libraries (e.g. Kubernetes, Go, etc.)
- Some version updates may be acceptable, for picking up bug fixes, but
  maintainers must exercise caution when reviewing.

There is possibility to have exceptions where breaking changes are allowed in
release branches. These are at the discretion of the project's maintainers, and
must be carefully considered before merging. An example of an allowed
breaking change might be a fix for a behavioral bug that was released in an
initial minor version (such as `v0.3.0`).


### Issue and Pull Request Management

Anyone may comment on issues and submit reviews for pull requests. However,
only accounts added as contributors can be assigned issues or pull requests.

Ubiquity maintainers can assign you an issue or pull request by leaving a
`/assign <your Github ID>` comment on the issue or pull request if you are a contributor.

### Commands and Workflow

Ubiquity follows the standard Kubernetes workflow: any PR
needs `lgtm` and `approved` labels, and PRs must pass the tests before being merged.
See [the contributor docs](https://github.com/kubernetes/community/blob/master/contributors/guide/pull-requests.md#the-testing-and-merge-workflow) for more info.

We use the same priority and kind labels as Kubernetes. See the labels
tab in GitHub for the full list.

## Release Process

Minor and patch releases are generally done immediately after a feature or
bugfix is landed, or sometimes a series of features tied together.

Minor releases will only be tagged on the *most recent* major release
branch, except in exceptional circumstances. Patches will be backported
to maintained stable versions, as needed.

Major releases will be done shortly after a breaking change is merged -- once
a breaking change is merged, the next release *must* be a major revision.
We don't intend to have a lot of these, so we may put off merging breaking
PRs until a later date.

### Exact Steps

Refer to the [releasing document](./docs/developers/releasing.md) for the exact steps.