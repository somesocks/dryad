# Adding a root

Now that we have a garden to work in, we should decide on our package structure.

To create a web server, we'll want the core server package, some tools, and some test cases.

We can start by creating the root for the core web server, by running `dryad root init ./dyd/roots/server`.

After creating the core web server, we need to set up the main script for the root, `./dyd/roots/server/dyd/commands/default`.  This script is what will be responsible for assembling a stem from the root during the build process.

Right now, we can leave it to be an empty shell script:

```
#!/usr/bin/env sh


```

After adding this line we can run `dryad garden build` and see output on the command line like so:

```
$:~/work/dryad/tutorial$ dryad garden build
[info] loading default scope: 
[info] dryad checking root server
[info] dryad building root server
[info] dryad done building root server
```

Note that running `dryad garden build` a second time does not trigger a new build of the root, since the build cache already has a record of a build for a root with those contents.

```
$:~/work/dryad/tutorial$ dryad garden build
[info] loading default scope: 
[info] dryad checking root server
```

Also note that the build created a new sprout, at `./dyd/sprouts/server`.  This is the resulting package from building the server root.  It has almost no content right now, just two files `dyd/fingerprint` and `dyd/traits/root-fingerprint`.

dryad copies no content from roots to stems by default, which means all assets or traits we want in a stem should come from us.  We can update `./dyd/roots/server/dyd/commands/default` to add a little more content during the build, like adding name and version traits to the stem.

```sh
#!/usr/bin/env sh

SRC_DIR=$DYD_STEM
DEST_DIR=$1

# add the package name as a trait
mkdir -p $DEST_DIR/dyd/traits
echo -n "hello-world-server" > $DEST_DIR/dyd/traits/name
echo -n "0.0.1" > $DEST_DIR/dyd/traits/version

```

After this, re-running the build will update the sprout and we should be able to see the two new trait files. 




