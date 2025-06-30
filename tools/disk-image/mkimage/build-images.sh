#!/bin/bash
# Copyright The Ubiquity Authors.
#
# Licensed under the Apache License, Version 2.0. Previously licensed under the Functional Source License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/ubiquitycluster/ubiquity/blob/main/LICENSE
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# This software was previously licensed under the Functional Source License but has now transitioned to an Apache 2.0 License
# as of June 2025.
# See the License for the specific language governing permissions and
# limitations under the License.

wd=$(pwd)
ELEMENTS_PATH="$wd/custom-elements:$wd/elements"

output_type=${output_type:-"qcow2"}
output_dir=${output_dir:-"$wd/output"}
input_dir=${input_dir:-"$wd/images"}
cache_dir=${cache_dir:-"$wd/cache"}
tmp_dir=${input_dir:-"$wd/tmp"}
log_dir=${log_dif:-"$wd/logs"}
image_filter=${image_filter:-""}
docker_tag=${docker_tag:-"latest"}
docker_repo=${docker_repo:-"ubiquity.azurecr.io"}

images=$(ls $input_dir | grep -E "$image_filter")
dib_cmd="ELEMENTS_PATH='$ELEMENTS_PATH' /usr/local/bin/disk-image-create -t $output_type --image-cache $cache_dir --checksum --logfile $log_dir"
echo images | xargs echo "images found:"

for img in $images; do
    if [[ $1 != "" ]]; then
      img=$1
      docker_target=""
      input_elements_file="$input_dir/$img/elements"
      input_env="$input_dir/$img/env"
      output_path="$output_dir/$img"
      echo "building $img from $input_elements_file and $input_env to create $output_path"
      input_elements=$(cat $input_elements_file | grep -v "#" | xargs echo)
      echo "image elements requested: $input_elements"
      build_cmd=". $input_env; $dib_cmd -o $output_path"
      if [[  "${output_type}" =~ "docker" ]]; then
        docker_target="${docker_repo}/${img}:${docker_tag}"
          build_cmd+=" --docker-target ${docker_target}"
      fi
      build_cmd+=" ${input_elements}"
      echo "Build command: $build_cmd"
      sudo -EH bash -c "$build_cmd"
      exit 0
    else
      docker_target=""
      input_elements_file="$input_dir/$img/elements"
      input_env="$input_dir/$img/env"
      output_path="$output_dir/$img"
      echo "building $img from $input_elements_file and $input_env to create $output_path"
      input_elements=$(cat $input_elements_file | grep -v "#" | xargs echo)
      echo "image elements requested: $input_elements"
      build_cmd=". $input_env; $dib_cmd -o $output_path"
      if [[  "${output_type}" =~ "docker" ]]; then
          docker_target="${docker_repo}/${img}:${docker_tag}"
          build_cmd+=" --docker-target ${docker_target}"
      fi
      build_cmd+=" ${input_elements}"
      echo "Build command: $build_cmd"
      sudo -EH bash -c "$build_cmd"
    fi 
done


