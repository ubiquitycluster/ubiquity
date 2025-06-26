# mkimage

A collection of diskimage builder (dib) elements and image. 

## Prepare Dev Environment

This repo provides a prep.sh scripts to init/update diskimage-buider as a submodule.
prep.sh will also install diskimage-builder and its dependencies. 
This has be minimally tested on Rocky Linux 8.7 and 9.3 but should also work on ubuntu 20.04 and 22.04.


## Building Images

To build all images run ``./build-images.sh``

To build a subset of images set the image_filter variable
e.g. ``image_filter="podm" ./build-images.sh``
``image_filter`` supports any regex supported by ``grep -E``

To build images of a specific format specify ``output_type``.
``output_type``  defaults to ``raw`` and can accept a comma
seperated list of values. The supported output_types values
depends on the version of diskimage-builder used.

Built images can be found in the output dir in the root of the repo.
The content of the  output, cache and tmp dirs are ingored by git.
The tmp dir is provided as a convient location to mount built images
and has no other fucntion. diskimage-build creates a seperate temp
dir when building images.

A cleanup script ``./cleanup-output.sh`` is provided to automatically
delete all images in the output dir. this script does not cleanup
built docker images and does not prompt the user for confirmation.
if you use it be warned that it will delete all your built images!!!

## Addtional Info

Building images requires sudo right to be able to create loopback devices
and mount/create filesystems. The build-images.sh script will invoke
diskimage-builder with sudo to allow it to build images without permision
issues.

This repo is a work in progress.
