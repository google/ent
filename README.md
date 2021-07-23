# 🌳 Ent

Ent (named after the [Ent](https://en.wikipedia.org/wiki/Ent) species from _The
Lord of the Rings_ by J. R. R. Tolkien) is an experimental universal, scalable,
general purpose, Content-Addressable Store (CAS) to explore verifiable data
structures, policies and graphs.

The end goal of Ent is for there to exist a few large Ent servers around the
world that store arbitrary content, and are connected with each other in a
federated way. Similarly to how git repositories are normally hosted on one of a
few large websites, arbitrary content would be stored on one or more Ent
servers, and users would interact with them via command line clients, web UI, or
libraries integrated in other applications.

Ent servers may store arbitrary static content, from source code, to binary
artifacts, to entire websites.

Currently, only static and public content is supported; content that is private
or restricted should not be put in Ent, since doing so would make it publicly
available forever.

Since everything in Ent is content-addressed, it is not necessary to blindly
trust any Ent server about the integrity of the data it provides; assuming the
hash of an object is obtained from a trustworthy source, the corresponding data
returned by an Ent server is checked by the client to match that particular
hash, effectively making the content self-verifying.

This is not an officially supported Google product.

## Object Store

At its core, Ent exposes a low-level Object Store API, which allows clients to
store and retrieve raw (uninterpreted) bytes (called objects) by their hash.

For instance, an object containing the string `ent` (in ASCII / UTF-8) is
identified by the following object hash (sha256):

`b86a048d168012ef5c3f960bd96646826915d5bce747bc239489e1832cb15c78`

## [DAG](https://en.wikipedia.org/wiki/Directed_acyclic_graph) Service

On top of the object store API, an Ent server may also expose a higher-level DAG
Service API.

Each Node is represented by an underlying object by its hash, but it is
interpreted differently according to its kind.

A Node id has the following
[multicodec](https://github.com/multiformats/multicodec) form:

`bafkreifynici2fuaclxvyp4wbpmwmrucnek5lphhi66chfej4gbszmk4pa`

Each Node may be of one of the following kinds:

- Raw Node (0x55): an uninterpreted byte sequence (i.e. the same as an Object
  from the underlying Object Store).
- DAG Node (0x70): an object in
  [DAG-protobuf](https://ipld.io/docs/codecs/known/dag-pb/) format, containing
  some data as bytes and zero or more links to other nodes, referenced by their
  node id.

Since the Node kind is part of the Node id, the DAG Service implementation
offers a more expressive API than the Object Store: it allows traversing the
DAG, even without knowing the semantics of the individual Nodes from an
application point of view.

For instance, a DAG Node may be the root of a tree that represents a file system
hierarchy; using the Object Store API, a client would have to read the root
Node, parse the links to the child nodes, and then retrieve these; since each
lookup may require a round trip to the Ent server, this may be slow and
inefficient. Using the DAG Service, the client may request the server to return
not only the root node itself, but also the content of any nodes that are linked
by it, recursively; the DAG Service would still perform individual lookups in
its Object Store, but in general this would be faster than letting the client do
it sequentially; the DAG Service then returns the list of Nodes reachable from
the root Node in a single batch.

More complex APIs may allow the client to also notify the server which sub-DAGs
(if any) it already has locally, so that the server only needs to send back the
additional nodes that the client does not already have (see also
https://github.blog/2015-09-22-counting-objects/#the-counting).

The goal of the DAG Service API is to make it possible to implement a variety of
higher-level protocols on top of it, which take advantage of the graph structure
of the data.

## Web Server

An Ent server may also expose a web API that serves Nodes over HTTP.

For each Node id, the Ent server hosts an entire website at a dedicated
subdomain origin that serves the entire DAG rooted at that Node.

For instance, assuming an Ent server hosted at `example.com`, the URL
`http://bafybeihzxousseuohykoj5ms2qc236sobkz6vdfnbrymxu7d4qo6qigi44.www.example.com/foo/bar`
corresponds to serving the result of fetching Node with id
`bafybeihzxousseuohykoj5ms2qc236sobkz6vdfnbrymxu7d4qo6qigi44`, interpreting it
as a DAG Node, following link named `foo`, interpreting that as DAG Node,
following link named `bar`, and interpreting the result as bytes to serve to the
browser. Note that in this case the browser would not validate the integrity of
any of the nodes involved, so it would have to trust the Ent server at that
particular domain to traverse the DAG correctly and serve the correct content.

### Examples

An individual file corresponds to a Raw Node, or, equivalently, to an Object,
whose content is the exact content of the file itself. Attributes and metadata
are not represented.

A directory corresponds to a DAG Node, with pointers to individual files or
directories it contains; the format of the data and links of this Node needs to
be agreed upon, and there are various possible alternatives; for instance, each
link may have the name of the file, and point to the content of that file; in
this way though it is not possible to represent metadata associated with that
file (e.g. whether the file is executable). A different (and incompatible)
representation format may store a list of entries containing name and other
metadata as data in the DAG Node, and have corresponding links pointing to the
content of each file.

A zip file, Docker image, or other archive format would be represented similarly
(if not identically) to a directory.

Most modern programming languages (Rust, Go, Javascript) reimplement their own
package manager to allow fetching dependencies. They usually use a combination
of a dependency specification file (e.g. Cargo.toml) which is manually edited,
and a lockfile (e.g. Cargo.lock) which is automatically generated by the
compiler given the specification file and the current state of the package
ecosystem and maps each dependency to a particular version of a package, usually
also including the hash of the content of the target package. With Ent, the role
of a package manager (in any language) would be purely to resolve package
specifications to individual versions, identified by their hash. Then, the
compiler may generate a lockfile compatible with Ent (or use Ent internally as a
library), but would not need to fetch the package content from a
language-specific server. Additionally, if both the source of the package and
its distributable version have a subset of the files or directories in common,
those objects would be literally shared by Ent as part of the DAG structure.

A git repository
[index](https://shafiul.github.io//gitbook/7_the_git_index.html) would map to an
Ent DAG (although the individual node hashes would differ from the ones that git
uses, if nothing else because git uses SHA1 hashes, but also because the format
of the metadata would differ too).

A Merkle tree such as [Trillian](https://github.com/google/trillian) would be
represented by a root DAG Node, with pointers to its children. Note that instead
of having to stop at the hashes of the content that is indexed, a Merkle tree on
Ent may actually extend to include the actual content itself in the lowest layer
of the tree, since these could be represented as Raw Nodes.

## Server

The Ent server exposes an object store API and a node service API.

In order to run the server locally, use the following command:

```bash
./run_server
```

## Command-Line Interface

The Ent CLI offers a way to operate on files on the local file system and sync
them to one or more Ent remotes via their node service API.

### Installation

The CLI can be built and installed with the following command:

```bash
go install ./cmd/ent
```

And is then available via the binary called `ent`:

```bash
ent help
```

### Configuration

The CLI relies on a local configuration file at `~/.config/ent.toml`, which
should contain a list of remotes, e.g.:

```toml
default_remote = "fs"

[remotes.fs]
path = "/tmp/ent"

[remotes.localhost]
url = "http://localhost:8080"

[remotes.obj]
url = "https://storage.googleapis.com/ent-objects"
```

Note that `~` and env variables are **not** expanded.

### `ent status`

Returns a summary of each file in the current directory, indicating for each of
them whether or not it is present in the remote.

### `ent push`

Pushes any file from the current directory to the remote if it is not already
there.

### `ent make`

Reads a file called `entplan.toml` in the current directory, such as the
following:

```toml
[[overrides]]
path = "tools/node"
from = "bafkreidhwdqb3p6lqzxe55na5hrdrw7d5meput4mpfcixcspo2nevbbafi"
executable = true

[[overrides]]
path = "tools/prettier"
from = "bafybeihnebwksdlbbimznfp7b2itgw6rzqg556vzpnxhdtekwx2ddbvbty"
```

Each `overrides` entry specifies a local path and the id of a node to pull into
that path from a remote.

For each entry, `ent make` creates the directory at the specified path (if not
already existing) and recursively pulls the specified node into it.

Directories not specified in `entplan.toml` are left unaffected.

It is conceptually similar to
[git submodules](https://git-scm.com/book/en/v2/Git-Tools-Submodules).

**TRY ME** (on Linux): To try this out this functionality in this repository,
after configuring the remotes as above and installing the `ent` CLI, run
`ent make --remote=obj` from the repository root; this will create the directory
`tools` and pull into it a specific immutable version of `NodeJS` (as a single
executable binary) and `prettier` (as a directory), as specified in
`entplan.toml`; after this, you can run `./format` to automatically format this
README using those specific versions of the tools that were just downloaded.
