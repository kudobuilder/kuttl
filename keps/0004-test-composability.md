---
kep-number: 4
short-desc: This KEP proposes a scheme for making test steps, asserts, and manifests composable in order to prevent repetition of YAML across test cases.
title: KUTTL Test Composability
authors:
  - "@jbarrick-mesosphere"
owners:
  - "@jbarrick-mesosphere"
  - "@kensipe"
editor: "@kensipe"
creation-date: 2020-02-12
last-updated: 2020-04-07
status: provisional
---

# KUTTL Test Composability

This KEP proposes a scheme for making test steps, asserts, and manifests composable in order to prevent repetition of YAML across test cases.

## Table of Contents

 * [Summary](#summary)
 * [Motivation](#motivation)
    * [Goals](#goals)
 * [Proposal](#proposal)
    * [Testing Life-cycle](#testing-life-cycle)
    * [Test Case file structure](#test-case-file-structure)

## Summary

KUTTL tests can tend to be quite repetitive, for example, for test case environment setup or if there are any common resources across test cases. Additionally, sometimes it is useful to compose common asserts.

## Motivation

This is KEP intends to improve the developer experience of writing tests, make tests easier to maintain and reduce duplication by making the files within them more composable. For example, one user [indicated that the tests are overly verbose](https://github.com/kudobuilder/kudo/issues/1311#issuecomment-580709826). Additionally, because this duplication can be expensive, some users deploy common resources into a single namespace breaking namespace isolation and it can make orchestrating tests difficult.

Beyond just composability, the magic file name scheme indicating step index and whether or not a file is an `assert` or `errors` file can be opaque and hard to understand for new users. The proposal in this KEP will also make test structure easier to understand.

### Goals

* Improve composability of test steps, manifests, and asserts.
* Improve reusability of YAML files.
* Provide better integration with existing repository structures (e.g., deploy operator YAMLs from a release directory).
* Make test file structure easier to understand.

## Proposal

The `TestStep` resource will be extended to allow users to provide manifest, errors, and assert files as a list. Three settings will be added: `apply` , `assert`, `error`. Each is a list of files that can either be a single file name or a directory to recurse over.

* `apply`: a list of files to apply for the test step.
* `assert`: a list of files containing test asserts to expect.
* `error`: a list of files containing test errors that will cause the test to fail if their state is observed.

An example use of this might look like:

```
apiVersion: kudo.dev/v1alpha1
kind: TestStep
apply:
- ./install.yaml
- ../../deploy/
assert:
- ../common/check-install.yaml
error:
- ./install-error.yaml
```

This will compose well with the existing `commands` feature of `TestSteps` (note: `commands` will be run prior to either `apply`, `assert`, or `error` during the actual test execution).

### Testing Life-cycle

Needed in order to make this `implementable` is definition of life-cycle with TestSteps, numbered test files, commands and controller hooks (controller setup).

### Test Case file structure

By default, `TestCases` will use the same file structure they did before (documented in KEP-0002), where the first part of the file name indicates the `TestStep's` index and file usage (`assert`, `errors`, or other for files to apply). This will maintain backwards compatibility and the new `TestStep` settings can be used on any existing tests.

However, users will be able to construct a `TestCase` from a series of `TestSteps` by including many `TestSteps` in a single file:

```
apiVersion: kudo.dev/v1alpha1
kind: TestStep
apply:
- ./install.yaml
- ../../deploy/
assert:
- ../common/check-install.yaml
error:
- ./install-error.yaml
---
apiVersion: kudo.dev/v1alpha1
kind: TestStep
apply:
- ./install.yaml
- ../../deploy/
assert:
- ../common/check-install.yaml
error:
- ./install-error.yaml
```

This file can be specified in place of the old test case directory format as a single YAML file, or, if it is placed in a test case directory, the files in the test case directory must still conform to the old test step index format to ensure proper ordering of `TestSteps` in multiple files.
