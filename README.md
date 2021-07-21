# Ent

Ent is an experimental universal, scalable, general purpose, Content-Addressable
Store (CAS) to explore verifiable data structures, policies and graphs.

The end goal of Ent is for there to exist a few large Ent servers around the
world that store arbitrary content, and are connected with each other in a
federated way. Similarly to how git repositories are normally hosted on one of a
few large websites, arbitrary content would be stored on one or more Ent
servers, and users would interact with them via command line clients, web UI, or
libraries integrated in other applications.

Ent servers may store arbitrary static content, from source code, to binary
artifacts, to entire websites.

Since everything in Ent is content-addressed, it is not necessary to blindly
trust any Ent server about the integrity of the data it provides; assuming the
hash of an object is obtained from a trustworthy source, the corresponding data
returned by an Ent server is checked by the client to match that particular
hash, effectively making the content self-verifying.

This is not an officially supported Google product.

## Object Store

At its core, Ent exposes a low-level object store API, which allows clients to
store and retrieve raw (uninterpreted) bytes (called objects) by their hash.

For instance, an object containing the string `ent` (in ASCII / UTF-8) is
identified by the following object hash (sha256):

`b86a048d168012ef5c3f960bd96646826915d5bce747bc239489e1832cb15c78`

## Node Service

On top of the object store API, an Ent server may also expose a higher-level
node service API.

Each node is represented by an underlying object by its hash, but it is
interpreted differently according to its kind.

A node id has the following
[multicodec](https://github.com/multiformats/multicodec) form:

`bafkreifynici2fuaclxvyp4wbpmwmrucnek5lphhi66chfej4gbszmk4pa`

Each node may be of one of the following kinds:

- raw node (0x55): an uninterpreted byte sequence.
- DAG node (0x70): a node in
  [DAG-protobuf](https://ipld.io/docs/codecs/known/dag-pb/) format with links to
  zero or more other nodes, referencing them by their node id.

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

### `status`

`ent status` returns a summary of each file in the current directory, indicating
for each of them whether or not it is present in the remote.

### `push`

`ent push` pushes any file from the current directory to the remote if it is not
already there.

### `make`

`ent make` reads a file called `entplan.toml` in the current directory, such as
the following:

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

**TRY ME**: To try this out this functionality in this repository, after
configuring the remotes as above, run `ent make --remote=obj` from the
repository root; this will create the directory `tools` and pull into it a
specific version of NodeJS and prettier, as per `entplan.toml`; after this, run
`./format` to automatically format this README using those specific versions of
the tools.
