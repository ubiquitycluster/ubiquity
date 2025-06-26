# EasyBuild (via EESSI)

As part of our installations, we include EasyBuild via EESSI. We do this by utilising CVMFS (the Cern VM FileSystem) within our environments to fuse-mount EESSI pre-built modulefiles (including EasyBuild) optimised for your architecture.

## General usage

On a login system, to use the EESSI software firstly you have to source the environment:

```
source /cvmfs/software.eessi.io/versions/2023.06/init/bash
```

Then you'll be presented with the ability to use lmod to either use the pre-compiled modules (that match against your architecture using archdetect) using `module avail`, `module load <module>` `module list`, and so on, or recompile yourself for your environment.
```
[ubiqsupport@login ~]$ source /cvmfs/software.eessi.io/versions/2023.06/init/bash
Found EESSI repo @ /cvmfs/software.eessi.io/versions/2023.06!
archdetect says x86_64/intel/skylake_avx512
Using x86_64/intel/skylake_avx512 as software subdirectory.
Found Lmod configuration file at /cvmfs/software.eessi.io/versions/2023.06/software/linux/x86_64/intel/skylake_avx512/.lmod/lmodrc.lua
Found Lmod SitePackage.lua file at /cvmfs/software.eessi.io/versions/2023.06/software/linux/x86_64/intel/skylake_avx512/.lmod/SitePackage.lua
Using /cvmfs/software.eessi.io/versions/2023.06/software/linux/x86_64/intel/skylake_avx512/modules/all as the directory to be added to MODULEPATH.
Using /cvmfs/software.eessi.io/host_injections/2023.06/software/linux/x86_64/intel/skylake_avx512/modules/all as the site extension directory to be added to MODULEPATH.
Found libcurl CAs file at RHEL location, setting CURL_CA_BUNDLE
Initializing Lmod...
Prepending /cvmfs/software.eessi.io/versions/2023.06/software/linux/x86_64/intel/skylake_avx512/modules/all to $MODULEPATH...
Prepending site path /cvmfs/software.eessi.io/host_injections/2023.06/software/linux/x86_64/intel/skylake_avx512/modules/all to $MODULEPATH...
Environment set up to use EESSI (2023.06), have fun!
{EESSI 2023.06} [ubiqsupport@login ~]$ module av

--------------------------------------------------------------- /cvmfs/software.eessi.io/versions/2023.06/software/linux/x86_64/intel/skylake_avx512/modules/all ----------------------------------------------------------------
   Abseil/20230125.2-GCCcore-12.2.0                 gfbf/2023a                                            libunwind/1.6.2-GCCcore-12.3.0                          PROJ/9.3.1-GCCcore-13.2.0                     (D)
   Abseil/20230125.3-GCCcore-12.3.0          (D)    gfbf/2023b                                  (D)       libunwind/1.6.2-GCCcore-13.2.0                   (D)    protobuf-python/4.24.0-GCCcore-12.3.0
   ALL/0.9.2-foss-2023a                             Ghostscript/10.0.0-GCCcore-12.2.0                     libvorbis/1.3.7-GCCcore-12.2.0                          protobuf/23.0-GCCcore-12.2.0
   AOFlagger/3.4.0-foss-2023b                       Ghostscript/10.01.2-GCCcore-12.3.0          (D)       libvorbis/1.3.7-GCCcore-12.3.0                   (D)    protobuf/24.0-GCCcore-12.3.0                  (D)
   archspec/0.2.1-GCCcore-12.3.0                    giflib/5.2.1-GCCcore-12.2.0                           libwebp/1.3.1-GCCcore-12.3.0                            PuLP/2.8.0-foss-2023a
   Armadillo/11.4.3-foss-2022b                      giflib/5.2.1-GCCcore-12.3.0                           libwebp/1.3.2-GCCcore-13.2.0                     (D)    pybind11/2.10.3-GCCcore-12.2.0
   Armadillo/12.6.2-foss-2023a                      giflib/5.2.1-GCCcore-13.2.0                 (D)       libxc/6.1.0-GCC-12.2.0                                  pybind11/2.11.1-GCCcore-12.3.0
   Armadillo/12.8.0-foss-2023b               (D)    git/2.38.1-GCCcore-12.2.0-nodocs                      libxc/6.2.2-GCC-12.3.0                           (D)    pybind11/2.11.1-GCCcore-13.2.0                (D)
   arpack-ng/3.8.0-foss-2022b                       git/2.41.0-GCCcore-12.3.0-nodocs                      libxml2/2.10.3-GCCcore-12.2.0                           pyfaidx/0.8.1.1-GCCcore-12.3.0
   arpack-ng/3.9.0-foss-2023a                       git/2.42.0-GCCcore-13.2.0                   (D)       libxml2/2.11.4-GCCcore-12.3.0                           PyOpenGL/3.1.7-GCCcore-12.3.0
   arpack-ng/3.9.0-foss-2023b                (D)    GitPython/3.1.40-GCCcore-12.3.0                       libxml2/2.11.5-GCCcore-13.2.0                    (D)    PyQt-builder/1.15.4-GCCcore-12.3.0
   arrow-R/11.0.0.3-foss-2022b-R-4.2.2              GLib/2.75.0-GCCcore-12.2.0                            libxslt/1.1.37-GCCcore-12.2.0                           PyQt5/5.15.10-GCCcore-12.3.0
   arrow-R/14.0.1-foss-2023a-R-4.3.2         (D)    GLib/2.77.1-GCCcore-12.3.0                            libxslt/1.1.38-GCCcore-12.3.0                           Pysam/0.22.0-GCC-12.3.0
   Arrow/11.0.0-gfbf-2022b                          GLib/2.78.1-GCCcore-13.2.0                  (D)       libxslt/1.1.38-GCCcore-13.2.0                    (D)    pystencils/1.3.4-gfbf-2023b
   Arrow/14.0.1-gfbf-2023a                   (D)    GLPK/5.0-GCCcore-12.2.0                               libyaml/0.2.5-GCCcore-12.3.0                            pytest-flakefinder/1.1.0-GCCcore-12.3.0
   ASE/3.22.1-gfbf-2022b                            GLPK/5.0-GCCcore-12.3.0                     (D)       libyaml/0.2.5-GCCcore-13.2.0                     (D)    pytest-rerunfailures/12.0-GCCcore-12.3.0
   at-spi2-atk/2.38.0-GCCcore-12.2.0                GMP/6.2.1-GCCcore-12.2.0                              LittleCMS/2.14-GCCcore-12.2.0                           pytest-shard/0.1.2-GCCcore-12.3.0
   at-spi2-atk/2.38.0-GCCcore-12.3.0         (D)    GMP/6.2.1-GCCcore-12.3.0                              LittleCMS/2.15-GCCcore-12.3.0                           Python-bundle-PyPI/2023.06-GCCcore-12.3.0
   at-spi2-core/2.46.0-GCCcore-12.2.0               GMP/6.3.0-GCCcore-13.2.0                    (D)       LittleCMS/2.15-GCCcore-13.2.0                    (D)    Python-bundle-PyPI/2023.10-GCCcore-13.2.0     (D)
   at-spi2-core/2.49.91-GCCcore-12.3.0       (D)    gmpy2/2.1.5-GCC-12.3.0                                LLVM/15.0.5-GCCcore-12.2.0                              python-casacore/3.5.2-foss-2023b
   ATK/2.38.0-GCCcore-12.2.0                        gmpy2/2.1.5-GCC-13.2.0                      (D)       LLVM/16.0.6-GCCcore-12.3.0                              python-isal/1.1.0-GCCcore-12.3.0
   ATK/2.38.0-GCCcore-12.3.0                 (D)    gnuplot/5.4.8-GCCcore-12.3.0                          LLVM/16.0.6-GCCcore-13.2.0                       (D)    Python/2.7.18-GCCcore-12.2.0-bare
```

## Compiling Local Modules from EESSI
Sometimes you will need to compile modules and software locally - Maybe your operating system is different, your GLIBC libraries are different and so on.

To do this, you can invoke EasyBuild from within EESSI, and build to a local directory.

You do this by creating an EasyBuild template configuration file, and pointing that configuration file to a shared directory and building versions of the software locally using EasyBuild Commands, issued below.

## Generating an EasyBuild template configuration file
Since EasyBuild v1.10, a command line option --confighelp is available that prints out the help text as an annotated configuration file. This can be used as an empty template configuration file:

```
mkdir -p $HOME/.config/easybuild
eb --confighelp > $HOME/.config/easybuild/config.cfg
$ head $HOME/.easybuild/config.cfg
[MAIN]
# Enable debug log mode (def False)
#debug=
# Enable info log mode (def False)
#info=
# Enable info quiet/warning mode (def False)
#quiet=

[basic]
# Print build overview incl. dependencies (full paths) (def False)
```
## Environment variables
All configuration settings listed as long options in `eb --help` can also be specified via `EASYBUILD_`-prefixed environment variables.

Configuration settings specified this way always override the corresponding setting specified in a configuration file.

For example, to enable debug logging using an environment variable:
```
export EASYBUILD_DEBUG=1
```
More examples of using environment variables to configure EasyBuild are shown in the sections below.

**Tip**

Any configuration option of EasyBuild which can be tuned by command line or via the configuration file, can also be tuned via a corresponding environment variable.

**Note**

If any `$EASYBUILD_`-prefixed environment variables are defined that do not correspond to a known configuration option, EasyBuild will report an error message and exit.

## Command line arguments
The configuration type with the highest precedence are the eb command line arguments, which override settings specified through environment variables or in configuration files.

For some configuration options, both short and long command line arguments are available (see `eb --help`); the long options indicate how the configuration setting should be specified in a configuration file or via an environment variable (`$EASYBUILD_<LONGOPTION>`).

For boolean configuration settings, both the `--<option>` and `--disable-<option>` variants are always available.

### Examples (more below):
enable debug logging (long option) and logging to stdout (short option)
```
eb --debug -l ...
```
use /dev/shm as build path, install to temporary install path, disable debug logging
```
eb --buildpath=/dev/shm --installpath=/tmp/$USER --disable-debug ...
```

### Overview of current configuration
(`--show-config`, `--show-full-config`)

To get an overview of the current EasyBuild configuration across all configuration types, you can use `eb --show-config`.

The output will specify:

- any configuration setting for which the current value is different from the default value
- a couple of selected important configuration settings (even if they are still set to the default value), i.e.:
- build path
- install path
- path to easyconfigs repository
- the robot search path
- source path
- through which configuration type each setting was defined
- i.e., default value, configuration file, environment variable or command line argument

### Example output
```
$ cat $HOME/.config/easybuild/config.cfg
[config]
buildpath = /tmp/eb-build

$ export EASYBUILD_MODULES_TOOL=Lmod
$ export EASYBUILD_OPTARCH=''

$ eb --show-config --installpath=$HOME/apps --job-cores=4
#
# Current EasyBuild configuration
# (C: command line argument, D: default value, E: environment variable, F: configuration file)
#
buildpath      (F) = /tmp/eb-build
installpath    (C) = /Users/example/apps
job-cores      (C) = 4
modules-tool   (E) = Lmod
optarch        (E) = ''
repositorypath (D) = /Users/example/.local/easybuild/ebfiles_repo
robot-paths    (D) = /Users/example/easybuild-easyconfigs/easybuild/easyconfigs
sourcepath     (D) = /Users/example/.local/easybuild/sources
```
For a full overview of the current configuration, including all configuration settings, see `eb --show-full-config`.

### Available configuration settings
To obtain a full and up-to-date list of available configuration settings, see `eb --help`. We refrain from listing all available configuration settings here, to avoid outdated documentation.

A couple of selected configuration settings are discussed below, in particular the mandatory settings.

### Mandatory configuration settings
A handful of configuration settings are mandatory, and should be provided using one of the supported configuration types.

The following configuration settings are currently mandatory (more details in the sections below):

- Source path (`--sourcepath`)
- Build path (`--buildpath`)
- Software and modules install path (`--installpath`, `--installpath-software`, `--installpath-modules`)
- Easyconfigs repository (`--repository`, `--repositorypath`)
- Logfile format (`--logfile-format`)

If any of these configuration settings is not provided in one way or another, EasyBuild will complain and exit.

In practice, all of these have reasonable defaults (see `eb --help` for the default settings).

**Note**

The mandatory path-related options can be tweaked collectively via --prefix, see Overall prefix path (--prefix) for more information.

#### Source path (--sourcepath)
default: $HOME/.local/easybuild/sources/ (determined via Overall prefix path (--prefix))

The sourcepath configuration setting specifies the parent path of the directory in which EasyBuild looks for software source and install files.

Looking for the files specified via the sources parameter in the .eb easyconfig file is done in the following order of preference:

<sourcepath>/<name>: a subdirectory determined by the name of the software package
<sourcepath>/<letter>/<name>: in the style of the easyblocks/easyconfigs directories: in a subdirectory determined by the first letter (in lower case) of the software package and by its full name
<sourcepath>: directly in the source path
Note that these locations are also used when EasyBuild looks for patch files in addition to the various easybuild/easyconfigs directories that are listed in the $PYTHONPATH.

You can specify multiple paths, separated with :, in which EasyBuild will look for sources, but only the first one will be used for downloading, so one needs to make sure at least the first path is writable by the user invoking eb.

#### Build path (`--buildpath`)
default: `$HOME/.local/easybuild/build/` (determined via Overall prefix path (`--prefix`))

The buildpath configuration setting specifies the parent path of the (temporary) directories in which EasyBuild builds its software packages.

Each software package is (by default) built in a subdirectory of the specified buildpath under <name>/<version>/<toolchain><versionsuffix>.

Note that the build directories are emptied and removed by EasyBuild when the installation is completed (by default).

**Tip**

Using /dev/shm as build path can significantly speed up builds, if it is available and provides a sufficient amount of space. Setting up the variable EASYBUILD_BUILDPATH in your shell startup files makes this default. However be aware that, fi., two parallel GCC builds may fill up /dev/shm !

#### Software and modules install path
(`--installpath`, `--installpath-software`, `--installpath-modules`)

defaults:

- software install path: `$HOME/.local/easybuild/software` (determined via Overall prefix path (`--prefix`) and `--subdir-software`)
- modules install path: `$HOME/.local/easybuild/modules/all` (determined via Overall prefix path (`--prefix`), `--subdir-modules` and `--suffix-modules-path`)

There are several ways in which the software and modules install path used by EasyBuild can be configured:

using the direct configuration options `--installpath-software` and `--installpath-modules` (see below)
via the parent install path configuration option `--installpath` (see below)
via the overall prefix path configuration option `--prefix` (see Overall prefix path (`--prefix`))

##### DIRECT OPTIONS
(`--installpath-software` and `--installpath-modules`)

default: (no default specified)

The `--installpath-software` and `--installpath-modules` configuration options (available since EasyBuild v2.1.0) allow to directly specify the software and modules install paths, respectively.

These configuration options have precedence over all of the other configuration options that relate to specifying the install path for software and/or modules (see below).

###### PARENT INSTALL PATH: `--INSTALLPATH`
default: (no default specified)

The `--installpath` configuration option specifies the parent path of the directories in which EasyBuild should install software packages and the corresponding module files.

The install path for software and modules specifically is determined by combining `--installpath` with `--subdir-software`, and combining `--installpath` with `--subdir-modules` and `--suffix-modules-path`, respectively.

For more information on these companion configuration options, see Software and modules install path subdirectories (`--subdir-software`, `--subdir-modules`, `--suffix-modules-path`).

###### FULL INSTALL PATH FOR SOFTWARE AND MODULE FILE
The full software and module install paths for a particular software package are determined by the active module naming scheme along with the general software and modules install paths specified by the EasyBuild configuration.

Both the software itself and the corresponding module file will be installed in a subdirectory of the corresponding install path named according to the active module naming scheme (default format: <name>/<version>-<toolchain><versionsuffix>). Additionally, symlinks to the actual module file are installed in a subdirectory of the modules install path named according to the value of the moduleclass easyconfig parameter.

For more information on the module naming scheme used by EasyBuild, see Active module naming scheme (`--module-naming-scheme`).

###### UPDATING $MODULEPATH¶
To make the modules generated by EasyBuild available, the `$MODULEPATH` environment variable must be updated to include the modules install path.

The recommended way to do this is to use the module use command. For example:
```
eb --installpath=$HOME/easybuild
module use $HOME/easybuild/modules/all
```
It is probably a good idea to add this to your (favourite) shell .rc file, e.g., `~/.bashrc`, and/or the `~/.profile` login scripts, so you do not need to adjust `$MODULEPATH` every time you start a new session.

**Note**

Updating `$MODULEPATH` is not required for EasyBuild itself, since eb updates `$MODULEPATH` itself at runtime according to the modules install path it is configured with.

#### Easyconfigs repository (`--repository`, `--repositorypath`)
default: FileRepository at $HOME/.local/easybuild/ebfiles_repo (determined via Overall prefix path (--prefix))

EasyBuild has support for archiving (tested) .eb easyconfig files. After successfully installing a software package using EasyBuild, the corresponding .eb file is uploaded to a repository defined by the repository and repositorypath configuration settings.

Currently, EasyBuild supports the following repository types (see also eb --avail-repositories):

- FileRepository('path', 'subdir'): a plain flat file repository; path is the path where files will be stored, subdir is an optional subdirectory of that path where the files should be stored
- GitRepository('path', 'subdir/in/repo': a non-empty bare git repository (created with `git init --bare` or `git clone --bare`); path is the path to the git repository (can also be a URL); subdir/in/repo is optional, and specifies a subdirectory of the repository where files should be stored in
- SvnRepository('path', 'subdir/in/repo'): an SVN repository; path contains the subversion repository location (directory or URL), the optional second value specifies a subdirectory in the repository
You need to set the repository setting inside a configuration file like this:
```
[config]
repository = FileRepository
repositorypath = <path>
```
Or, optionally an extra argument representing a subdirectory can be specified, e.g.:
```
export EASYBUILD_REPOSITORY=GitRepository
export EASYBUILD_REPOSITORYPATH=<path>,<subdir>
```
You do not have to worry about importing these classes, EasyBuild will make them available to the configuration file.

Using git requires the GitPython Python modules, using svn requires the pysvn Python module (see Dependencies).

If access to the easyconfigs repository fails for some reason (e.g., no network or a missing required Python module), EasyBuild will issue a warning. The software package will still be installed, but the (successful) easyconfig will not be automatically added to the archive (i.e., it is not considered a fatal error).

#### Logfile format (`--logfile-format`)
default: easybuild, easybuild-%(name)s-%(version)s-%(date)s.%(time)s.log

The logfile format configuration setting contains a tuple specifying a log directory name and a template log file name. In both of these values, using the following string templates is supported:

- %(name)s: the name of the software package to install
- %(version)s: the version of the software package to install
- %(date)s: the date on which the installation was performed (in YYYYMMDD format, e.g. 20120324)
- %(time)s: the time at which the installation was started (in HHMMSS format, e.g. 214359)

**Note**

Because templating is supported in configuration files themselves (see Templates and constants supported in configuration files), the '%' character in these template values must be escaped when used in a configuration file (and only then), e.g., '%%(name)s'. Without escaping, an error like InterpolationMissingOptionError: Bad value substitution will be thrown by ConfigParser.

For example, configuring EasyBuild to generate a log file mentioning only the software name in a directory named easybuild can be done via the --logfile-format command line option:
```
eb --logfile-format="easybuild,easybuild-%(name)s.log" ...
```
or the `$EASYBUILD_LOGFILE_FORMAT` environment variable:

```
export EASYBUILD_LOGFILE_FORMAT="easybuild,easybuild-%(name)s.log"
```
or by including the following in an EasyBuild configuration file (note the use of '%%' to escape the name template value here):
```
logfile-format = easybuild,easybuild-%%(name)s.log
```
#### Optional configuration settings
The subsections below discuss a couple of commonly used optional configuration settings.

##### Overall prefix path (`--prefix`)
default: $HOME/.local/easybuild

The overall prefix path used by EasyBuild can be specified using the --prefix configuration option.

This affects the default value of several configuration options:

- source path
- build path
- software and modules install path
- easyconfigs repository path
- package path
- container path

##### Software and modules install path subdirectories
(`--subdir-software`, `--subdir-modules`, `--suffix-modules-path`)

defaults:

- software install path subdirectory (`--subdir-software`): software
- modules install path subdirectory (`--subdir-modules`): modules
- modules install path suffix (`--suffix-modules-path`): all

The subdirectories for the software and modules install paths (relative to --installpath, see install path) can be specified using the corresponding dedicated configuration options (available since EasyBuild v1.14.0).

For example:

```
export EASYBUILD_SUBDIR_SOFTWARE=installs
eb --installpath=$HOME/easybuild --subdir-modules=module_files ...
```
##### Modules tool (`--modules-tool`)
default: Lmod

Specifying the modules tool that should be used by EasyBuild can be done using the modules-tool configuration setting. A list of supported modules tools can be obtained using eb --avail-modules-tools.

Currently, the following modules tools are supported:

- Lmod (default): Lmod, an modern alternative to environment modules, written in Lua (lmod)
- EnvironmentModules: modern Tcl-only version of environment modules (4.x) (modulecmd.tcl)
- EnvironmentModulesC: Tcl/C version of environment modules, usually version 3.2.10 (modulecmd)
- EnvironmentModulesTcl: (ancient) Tcl-only version of environment modules (modulecmd.tcl)

You can determine which modules tool you are using by checking the output of type -f module (in a bash shell), or alias module (in a tcsh shell).

The actual module command (i.e., modulecmd, modulecmd.tcl, lmod, ...) must be available via $PATH (which is not standard), except when using Lmod (in that case the lmod binary can also be located via $LMOD_CMD) or when using Environment Modules (in that case the modulecmd.tcl binary can also be located via $MODULES_CMD).

For example, to indicate that EasyBuild should be using Lmod as modules tool:
```
eb --modules-tool=Lmod ...
```
##### Active module naming scheme (`--module-naming-scheme`)
default: EasyBuildModuleNamingScheme

The module naming scheme that should be used by EasyBuild can be specified using the module-naming-scheme configuration setting.
```
eb --module-naming-scheme=HierarchicalMNS ...
```
For more details, see the dedicated page on using a custom module naming scheme.

##### Module files syntax (`--module-syntax`)
default: Lua

supported since: EasyBuild v2.1

The syntax to use for generated module files can be specified using the --module-syntax configuration setting.

Possible values are:

- Lua: generate module files in Lua syntax: this requires the use of Lmod as a modules tool to consume the module files (see modules tool)
module file names will have the .lua extension
- Tcl: generate module files in Tcl syntax: Tcl module files can be consumed by all supported modules tools. Module files will contain a header string #%Module indicating that they are composed in Tcl syntax

**Note**

Lmod is able to deal with having module files in place in both Tcl and Lua syntax. When a module file in Lua syntax (i.e., with a .lua file name extension) is available, a Tcl module file with the same name will be ignored. The Tcl-based environment modules tool will simply ignore module files in Lua syntax, since they do not contain the header string that is included in Tcl module files.

**Note**

Using module files in Lua syntax has the advantage that Lmod does not need to translate from Lua to Tcl internally when processing the module files, which benefits responsiveness of Lmod when used interactively by users. In terms of Lmod-specific aspects of module files, the syntax of the module file does not matter; Lmod-specific statements supported by EasyBuild can be included in Tcl module files as well, by guarding them by a condition that only evaluates positively when Lmod is consuming the module file, i.e. 'if { [ string match "*tcl2lua.tcl" $env(_) ] } { ... }'. Only conditional load statements like 'load(atleast("gcc","4.8"))' can only be used in Lua module files.
