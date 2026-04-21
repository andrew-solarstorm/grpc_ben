const { PublicKey } = require('@solana/web3.js');
const bs58 = require('bs58').default || require('bs58'); // bs58 export handling
const { TokenProgramID } = require('./types'); // Wait, TokenProgram is usually in @solana/web3.js, let's hardcode the public key

const TOKEN_PROGRAM_ID = new PublicKey("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA");

class TransactionInfo {
    constructor() {
        this.Signature = null;
        this.Transaction = null;
        this.Meta = null;
        this.Slot = 0;
    }

    decodeFromBlock(tx, slot) {
        this.Signature = bs58.encode(tx.signature);
        this.Transaction = this.decodeTransaction(tx.transaction);
        this.Meta = this.decodeTransactionMeta(tx.meta);
        this.Slot = slot;
    }

    decodeTransaction(tx) {
        if (!tx) return null;
        const sigs = (tx.signatures || []).map(sig => bs58.encode(sig));
        return {
            Signatures: sigs,
            Message: this.decodeTransactionMessage(tx)
        };
    }

    decodeTransactionMessage(tx) {
        if (!tx || !tx.message) return null;
        const accountKeys = (tx.message.accountKeys || []).map(a => new PublicKey(a));

        const instructions = (tx.message.instructions || []).map(ix => this.decodeInstruction(ix));

        return {
            AccountKeys: accountKeys,
            Header: {
                NumRequiredSignatures: tx.message.header?.numRequiredSignatures || 0,
                NumReadonlySignedAccounts: tx.message.header?.numReadonlySignedAccounts || 0,
                NumReadonlyUnsignedAccounts: tx.message.header?.numReadonlyUnsignedAccounts || 0,
            },
            RecentBlockhash: tx.message.recentBlockhash ? bs58.encode(tx.message.recentBlockhash) : "",
            Instructions: instructions,
            AddressTableLookups: this.decodeLookupTables(tx)
        };
    }

    decodeInstruction(ix) {
        const accounts = [];
        if (ix.accounts) {
            for (let i = 0; i < ix.accounts.length; i++) {
                accounts.push(ix.accounts[i]);
            }
        }
        return {
            ProgramIDIndex: ix.programIdIndex,
            Accounts: accounts,
            Data: ix.data // Buffer or Uint8Array
        };
    }

    decodeLookupTables(tx) {
        if (!tx || !tx.message || !tx.message.addressTableLookups) return [];
        return tx.message.addressTableLookups.map(lut => ({
            AccountKey: new PublicKey(lut.accountKey),
            WritableIndexes: Array.from(lut.writableIndexes || []),
            ReadonlyIndexes: Array.from(lut.readonlyIndexes || [])
        }));
    }

    decodeTransactionMeta(meta) {
        if (!meta) return null;

        const preTb = (meta.preTokenBalances || []).map(b => this.decodeTokenBalance(b));
        const postTb = (meta.postTokenBalances || []).map(b => this.decodeTokenBalance(b));

        const rewards = (meta.rewards || []).map(r => this.decodeReward(r));

        const innerInstructions = (meta.innerInstructions || []).map(ix => this.decodeInnerInstruction(ix));

        const loadedAddresses = {
            Writable: (meta.loadedAddresses?.writable || []).map(k => new PublicKey(k)),
            Readonly: (meta.loadedAddresses?.readonly || []).map(k => new PublicKey(k))
        };

        return {
            Err: meta.err ? JSON.stringify(meta.err) : "",
            Fee: meta.fee,
            PreBalances: meta.preBalances || [],
            PostBalances: meta.postBalances || [],
            InnerInstructions: innerInstructions,
            PreTokenBalances: preTb,
            PostTokenBalances: postTb,
            LogMessages: meta.logMessages || [],
            Rewards: rewards,
            LoadedAddresses: loadedAddresses,
        };
    }

    decodeTokenBalance(b) {
        const owner = b.owner ? new PublicKey(b.owner) : null;
        const mint = new PublicKey(b.mint);

        return {
            AccountIndex: b.accountIndex,
            Owner: owner,
            Mint: mint,
            UiTokenAmount: {
                Amount: b.uiTokenAmount?.amount || "0",
                Decimals: b.uiTokenAmount?.decimals || 0,
                UiAmountString: b.uiTokenAmount?.uiAmountString || "0",
            }
        };
    }

    decodeInnerInstruction(b) {
        const ixs = (b.instructions || []).map(ix => {
            const accounts = [];
            if (ix.accounts) {
                for (let i = 0; i < ix.accounts.length; i++) {
                    accounts.push(ix.accounts[i]);
                }
            }
            return {
                ProgramIDIndex: ix.programIdIndex,
                Accounts: accounts,
                Data: ix.data
            };
        });

        return {
            Index: b.index,
            Instructions: ixs
        };
    }

    decodeReward(b) {
        return {
            Pubkey: new PublicKey(b.pubkey),
            Lamports: b.lamports,
            PostBalance: b.postBalance,
            RewardType: b.rewardType,
        };
    }
}

class TxExtractArgs {
    constructor() {
        this.AccountKeys = [];
        this.Instructions = [];
        this.Transaction = null;
        this.Prebalances = {};
        this.Postbalances = {};
        this.AccToMintMap = {}; // Maps Base58 string to PublicKey
        this.AccToOwnerMap = {}; // Maps Base58 string to PublicKey
        this.TxContext = null;
    }
}

function processTx(tx) {
    if (!tx || !tx.Transaction || !tx.Transaction.Message) return new TxExtractArgs();

    let accountKeys = [...tx.Transaction.Message.AccountKeys];
    if (tx.Meta && tx.Meta.LoadedAddresses) {
        accountKeys.push(...tx.Meta.LoadedAddresses.Writable);
        accountKeys.push(...tx.Meta.LoadedAddresses.Readonly);
    }

    let instructions = [];

    for (const ix of tx.Transaction.Message.Instructions) {
        const programID = accountKeys[ix.ProgramIDIndex];
        if (programID && programID.equals(TOKEN_PROGRAM_ID)) {
            instructions.push({
                ProgramIDIndex: ix.ProgramIDIndex,
                Accounts: ix.Accounts,
                Data: ix.Data // in Node.js this is already Uint8Array/Buffer from gRPC
            });
        }
    }

    const accToMintMap = {};
    const accToOwnerMap = {};

    if (tx.Meta && tx.Meta.InnerInstructions) {
        for (const inner of tx.Meta.InnerInstructions) {
            for (const ix of inner.Instructions) {
                const programID = accountKeys[ix.ProgramIDIndex];
                if (programID && programID.equals(TOKEN_PROGRAM_ID)) {
                    instructions.push({
                        ProgramIDIndex: ix.ProgramIDIndex,
                        Accounts: ix.Accounts,
                        Data: ix.Data
                    });
                }
            }
        }
    }

    const preBalances = {};
    const postBalances = {};

    if (tx.Meta) {
        for (const bal of (tx.Meta.PreTokenBalances || [])) {
            preBalances[bal.AccountIndex] = bal;
            if (bal.AccountIndex >= accountKeys.length) continue;
            const keyBase58 = accountKeys[bal.AccountIndex].toBase58();
            accToMintMap[keyBase58] = bal.Mint;
            if (bal.Owner) {
                accToOwnerMap[keyBase58] = bal.Owner;
            }
        }

        for (const bal of (tx.Meta.PostTokenBalances || [])) {
            postBalances[bal.AccountIndex] = bal;
            if (bal.AccountIndex >= accountKeys.length) continue;
            const keyBase58 = accountKeys[bal.AccountIndex].toBase58();
            accToMintMap[keyBase58] = bal.Mint;
            if (bal.Owner) {
                accToOwnerMap[keyBase58] = bal.Owner;
            }
        }
    }

    const args = new TxExtractArgs();
    args.AccountKeys = accountKeys;
    args.Instructions = instructions;
    args.Transaction = tx;
    args.Prebalances = preBalances;
    args.Postbalances = postBalances;
    args.AccToMintMap = accToMintMap;
    args.AccToOwnerMap = accToOwnerMap;

    return args;
}

module.exports = {
    TransactionInfo,
    TxExtractArgs,
    processTx
};
