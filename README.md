# What's the Point
The argument being made here is that you don't need to index the whole blockchain to detect spends from unauthorized parties. You can just search for the public key bytes in a streamed block or even within transaction being send around. This can be a service which exchanges pay for to avoid running a full node while operating an SPV base Wallet / treasury management platform.

The expectation being exchanges use P2PKH and the public key is secret until the utxo is spent, and if using SPV then the publicKey should thereafter never be used again.

## Next Steps
Probably add a bloom filter or something similar to allow checking large sets of public keys simultaneously. Although on this laptop this code already covers a 4GB files in 1 second, so for checking a few cold wallets this is already fast enough.