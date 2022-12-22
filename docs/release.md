# Release

> **Note: This document is only relevant for maintainers.**


This document will provide documentation for creating new releases

## Branches

In order to preserve the ability to apply security and bug fixes in a backward
compatible way this project uses release branches. Those branches have to follow
a naming schema:

```sh
git checkout -b release-<major>.<minor>
```

Whenever you want to create a new major or minor version of kubepost you have to
create a new release branch.

## Tagging

After creating a release branch based of the `master` branch a new git tag has
to be created. This tag will be used for the corresponding Github release and
a Docker tag, that will be applied to the image, that will be created within the workflow.

Git tags have to comply to a naming schema:

```sh
git tag v<major>.<minor>.<patch>[-rc.<counter>]
```

> Note: The counter mentioned above has to start at 1.

## Preparing a release

1. As the tagged versions should contain an up to date version of the changelog
   it is mandatory to run `make generate`. This will update the current changelog.

   > Note: Make sure, that all issues are closed, that are included within the release.

2. After generating the latest changelog the Readme has to be updated to contain
   the git tag, that will be used for the upcoming release.

3. Commit the changes and tag the commit with the tag you have written into the Readme
   within step 2.

## Follow up

After the release has been created it is often necessary to update the Readme within
the master branch with the latest available tag and changelog.

1. Check out the `master` branch

2. Create a `docs` feature branch

3. Generate the changelogs
   ```sh
   make generate-changelog
   ```

4. Update the `Quickstart` section within the Readme to point to the latest stable release.

5. Push your changes

6. Create a PR to merge the changes back into master.
    
