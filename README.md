ERC-5189 Mempool Bundler
==================================================

## Work in progress

This project is in the early stages of development and is not yet ready for production use.

## Overview

The project is a mempool transaction bundler for general purpose "operations". The project uses ERC-5189 as the reference standard for how the operations are defined and how they should be handled. The project includes a built-in sender, but it is designed to be used with a separate sender (or block builder).

## Usage

1. Create a copy of the `/etc/bundler-node.conf.sample` file and name it `/etc/bundler-1.conf`.

2. (Optional) Generate a random 12-word mnemonic and put it in the `mnemonic` field in the `/etc/bundler-1.conf` file.

3. Define the number of senders to run `num_senders` in the `/etc/bundler-1.conf` file.

4. Run with `make run`.

## Consuming the API

The API can be consumed using the client that can be found in the `/proto/client` directory. Note that the API is not yet stable and is subject to change.
