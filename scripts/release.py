import os
import re
import subprocess

import yaml
from yaml.loader import SafeLoader


def normalize_docker_tag(git_tag):
    docker_tag = re.sub('[^a-zA-Z0-9_.-]', '-', git_tag)
    docker_tag = re.sub('^v', '', docker_tag)
    return docker_tag


def compute_release_kustomization(kustomization_path, git_tag):
    with open(kustomization_path, 'r') as file:
        data = yaml.load(file, Loader=SafeLoader)

        docker_tag = normalize_docker_tag(git_tag)

        patch = [
            {
                "name": "controller",
                "newName": "ghcr.io/orbatschow/kubepost",
                "newTag": docker_tag
            }
        ]

        data['images'] = patch
        file.close()

        return data


def write_release_kustomization(kustomization_path, data):
    with open(kustomization_path, 'w') as file:
        yaml.dump(data, file)
        file.close()


def main():
    # get the repository directory
    repo_dir = subprocess.Popen(['git', 'rev-parse', '--show-toplevel'], stdout=subprocess.PIPE).communicate()[0].rstrip().decode('utf-8')

    # get the current git tag
    git_tag = subprocess.Popen(['git', 'name-rev', '--name-only', '--tags', 'HEAD'], stdout=subprocess.PIPE).communicate()[0].rstrip().decode('utf-8')

    # kustomization path
    kustomization_path = os.path.join(repo_dir, 'build', 'config', 'default', 'kustomization.yaml')

    # compute the docker tag

    # compute changes for root kustomization
    data = compute_release_kustomization(kustomization_path, git_tag)

    # write release tags within the root kustomization
    write_release_kustomization(kustomization_path, data)


if __name__ == "__main__":
    main()
