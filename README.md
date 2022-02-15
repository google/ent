# 🌳 Ent

Ent (named after the [Ent](https://en.wikipedia.org/wiki/Ent) species from _The
Lord of the Rings_ by J. R. R. Tolkien) is an experimental universal, scalable,
general purpose, Content-Addressable Store (CAS) to explore verifiable data
structures, policies and graphs.

This is not an officially supported Google product.

## Content-Addressability

Ent encourages a model in which files are referred to by their digest (as a
proxy for their content), instead of which server they happen to be located
(which is what a URL normally is for).

For example, instead of referring to the image below by its URL
`https://upload.wikimedia.org/wikipedia/commons/thumb/4/48/The_Calling_of_Saint_Matthew-Caravaggo_%281599-1600%29.jpg/405px-The_Calling_of_Saint_Matthew-Caravaggo_%281599-1600%29.jpg`,
in Ent it would be referred to by its digest
`sha256:f3e737f4d50fbf6bb6053e3b8c72d6bf7f1a7229aacf2e9b4c97e9dd27cb1dcf`.

![](https://upload.wikimedia.org/wikipedia/commons/thumb/4/48/The_Calling_of_Saint_Matthew-Caravaggo_%281599-1600%29.jpg/405px-The_Calling_of_Saint_Matthew-Caravaggo_%281599-1600%29.jpg)

The digest of a file is a stable cryptographic identifier for the actual data
that is contained in the file, and does not depend on which server the file
happens to be hosted on, or at what path. If at some point in the future the
file were to disappear from the original location and be made available at a
different location, the original URL would stop working, but the digest of the
file would remain the same, and can be used to refer to that file forever.

Additionally, using a digest to refer to a file is useful for security and
trustworthiness: if someone sends you the digest of a file to download (e.g. a
program to install on your computer), you can be sure that, by resolving that
digest to an actual file via Ent, the resulting file is exactly the one that the
sender intended, without having to trust the Ent Server, the Ent Index or the
server where the file is ultimately hosted.

## Installation

The Ent CLI can be built and installed with the following command, after having
cloned this repository locally:

```bash
go install ./cmd/ent
```

And is then available via the binary called `ent`:

```bash
ent help
```

## Examples

In order to fetch a file with a given digest, the `ent get` subcommand can be
used.

You can try the following command in your terminal, which fetches the text of
the _Treasure Island_ book:

```console
$ ent get sha256:4c350163715b7b1d0fc3bcbf11bfffc0cf2d107f69253f237111a7480809e192 | head
The Project Gutenberg EBook of Treasure Island, by Robert Louis Stevenson

This eBook is for the use of anyone anywhere in the United States and most
other parts of the world at no cost and with almost no restrictions
whatsoever.  You may copy it, give it away or re-use it under the terms of
the Project Gutenberg License included with this eBook or online at
www.gutenberg.org.  If you are not located in the United States, you'll have
to check the laws of the country where you are located before using this ebook.

Title: Treasure Island
```

The Ent CLI queries the default Ent index
(https://github.com/tiziano88/ent-index) to resolve the digest to a URL, and
then fetches the file at that URL, and also verifies that it corresponds to the
expected digest. It first buffers the entire file internally in order to verify
its digest, and only prints it to stdout if it does match the expected digest.

You can also manually double check that the returned file does in fact
correspond to the expected digest:

```console
$ ent get sha256:4c350163715b7b1d0fc3bcbf11bfffc0cf2d107f69253f237111a7480809e192 | sha256sum
4c350163715b7b1d0fc3bcbf11bfffc0cf2d107f69253f237111a7480809e192  -
```

## Ent Server

An Ent Server provides access to an underlying Ent store via an HTTP-based REST
API.

An Ent Server may be running locally (on port 27333 by default), or remotely.

Some Ent Servers require the user to be authenticated in order for the user to
read and / or write, which is performed via an API key.

## Ent Index

An Ent index is a "cheap" way to provide access to existing location-addressed
content on the internet, but in a content-addressable way.

It consists of a static website, which serves an entry for each digest, listing
one or more "traditional" URLs which may provide the file in question.

For instance, it may be serialized as a Git repository with a directory
structure corresponding to the digest of each entry, and a JSON file for each
entry that lists one or more URLs at which the object may be found.

The directory path is obtained by grouping sets of two digits from the digest,
and creating a nested folder for each of them; this is in order to limit the
number of files or directories inside each directory, since that would otherwise
not scale when there are millions of entries in the index.

For instance, the file with digest
`sha256:4c350163715b7b1d0fc3bcbf11bfffc0cf2d107f69253f237111a7480809e192` is
stored in the Ent index under the file
`/sha256/4c/35/01/63/71/5b/7b/1d/0f/c3/bc/bf/11/bf/ff/c0/cf/2d/10/7f/69/25/3f/23/71/11/a7/48/08/09/e1/92/entry.json`,
which contains the following entry:

https://github.com/tiziano88/ent-index/blob/fddaa4b78ec4f4ba1e2c1e3e1c0b5ae9b06565e2/sha256/4c/35/01/63/71/5b/7b/1d/0f/c3/bc/bf/11/bf/ff/c0/cf/2d/10/7f/69/25/3f/23/71/11/a7/48/08/09/e1/92/entry.json#L1

Note that the Ent index only stores URLs, not actual data, under the
_assumption_ that each URL will keep pointing to the same file forever.

The client querying the index is responsible to verify that the target file
still corresponds to the expected digest; if this validation fails, it means
that the URL was moved to point to a diferent file after it was added to the Ent
index.

### Updating the index

Currently, entries may be added to the default index by creating a comment in
https://github.com/tiziano88/ent-index/issues/1 containing the URL of the file
to index. A GitHub actions is then triggered that fetches the file, creates an
appropriate entry in the index, and commits that back into the git repository.

You can try this by picking a URL of an existing file, and creating a comment in
https://github.com/tiziano88/ent-index/issues/1 ; after a few minutes, the
GitHub action should post another comment in reply, confirming that the entry
was correctly incorporated in the index, and printing out its digest, which may
then be used with the Ent CLI as above.

If the URL stops pointing to the file that was originally indexed, the Ent CLI
will detect that and produce an error.

There is no process for cleaning up / fixing inconsistent entries in the index
(yet).

## Comparison with other systems

### IPFS

https://ipfs.io/

IPFS aims to be a fully decentralized and censorship-resistant protocol and
ecosystem, which heavily relies on content-addressability.
