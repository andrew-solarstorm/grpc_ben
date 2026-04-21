const { PublicKey } = require('@solana/web3.js');
const base58 = require('bs58').default || require('bs58'); // Handle both ESM and CJS exports

const TOKEN_PROGRAM_ID_STR = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA";

class TransactionInfo {
    constructor() {
        this.Signature = null;
        this.Transaction = null;
        this.Meta = null;
        this.Slot = 0;
    }

    decodeFromBlock(tx, slot) {
        this.Signature = tx.signature ? base58.encode(tx.signature) : "";
        this.Transaction = this.decodeTransaction(tx.transaction);
        this.Meta = this.decodeTransactionMeta(tx.meta);
        this.Slot = slot;
    }

    decodeTransaction(tx) {
        if (!tx) return null;
        const sigs = (tx.signatures || []).map(sig => base58.encode(sig));
        return {
            Signatures: sigs,
            Message: this.decodeTransactionMessage(tx),
            Signature: sigs.length > 0 ? sigs[0] : ""
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
                NumReadonlyUnsignedAccounts: tx.message.header?.numReadonlyUnsignedAccounts || 0
            },
            RecentBlockhash: tx.message.recentBlockhash ? base58.encode(tx.message.recentBlockhash) : "",
            Instructions: instructions,
            AddressTableLookups: this.decodeLookupTables(tx)
        };
    }

    decodeInstruction(ix) {
        return {
            ProgramIDIndex: ix.programIdIndex,
            Accounts: this.readUint8As16Slice(ix.accounts),
            Data: ix.data // Uint8Array equivalent
        };
    }

    decodeLookupTables(tx) {
        return (tx.message.addressTableLookups || []).map(lut => ({
            AccountKey: new PublicKey(lut.accountKey),
            WritableIndexes: Array.from(lut.writableIndexes || []),
            ReadonlyIndexes: Array.from(lut.readonlyIndexes || [])
        }));
    }

    decodeTransactionMeta(txMeta) {
        if (!txMeta) return null;
        const preTb = (txMeta.preTokenBalances || []).map(b => this.decodeTokenBalance(b));
        const postTb = (txMeta.postTokenBalances || []).map(b => this.decodeTokenBalance(b));
        const rewards = (txMeta.rewards || []).map(r => this.decodeReward(r));
        const innerInstructions = (txMeta.innerInstructions || []).map(ix => this.decodeInnerInstruction(ix));

        return {
            Err: txMeta.err ? JSON.stringify(txMeta.err) : "",
            Fee: txMeta.fee,
            PreBalances: Array.from(txMeta.preBalances || []),
            PostBalances: Array.from(txMeta.postBalances || []),
            InnerInstructions: innerInstructions,
            PreTokenBalances: preTb,
            PostTokenBalances: postTb,
            LogMessages: txMeta.logMessages || [],
            Rewards: rewards,
            LoadedAddresses: {} 
        };
    }

    decodeTokenBalance(b) {
        let mint = null;
        let owner = null;
        try { if (b.mint) mint = new PublicKey(b.mint); } catch(e) {}
        try { if (b.owner) owner = new PublicKey(b.owner); } catch(e) {}

        return {
            AccountIndex: b.accountIndex,
            Owner: owner,
            Mint: mint,
            UiTokenAmount: {
                Amount: b.uiTokenAmount?.amount || "",
                Decimals: b.uiTokenAmount?.decimals || 0,
                UiAmountString: b.uiTokenAmount?.uiAmountString || ""
            }
        };
    }

    decodeInnerInstruction(b) {
        const ixs = (b.instructions || []).map(ix => ({
            ProgramIDIndex: ix.programIdIndex,
            Accounts: this.readUint8As16Slice(ix.accounts),
            Data: ix.data
        }));
        return {
            Index: b.index,
            Instructions: ixs
        };
    }

    decodeReward(b) {
        let pk = null;
        try { if (b.pubkey) pk = new PublicKey(b.pubkey); } catch(e) {}
        return {
            Pubkey: pk,
            Lamports: b.lamports,
            PostBalance: b.postBalance,
            RewardType: b.rewardType
        };
    }

    readUint8As16Slice(data) {
        if (!data) return [];
        return Array.from(data);
    }
}

class TxExtractArgs {
    constructor() {
        this.AccountKeys = [];
        this.Instructions = [];
        this.Transaction = null;
        this.Prebalances = {};
        this.Postbalances = {};
        this.AccToMintMap = {}; 
        this.AccToOwnerMap = {}; 
        this.TxContext = null;
    }
}

function processTx(tx) {
    if (!tx || !tx.Transaction || !tx.Transaction.Message) {
        return new TxExtractArgs();
    }

    const accountKeys = tx.Transaction.Message.AccountKeys;
    let instructions = [];

    for (const ix of tx.Transaction.Message.Instructions) {
        const programID = accountKeys[ix.ProgramIDIndex];
        if (programID && programID.toBase58() === TOKEN_PROGRAM_ID_STR) {
            instructions.push({
                ProgramIDIndex: ix.ProgramIDIndex,
                Accounts: ix.Accounts,
                Data: ix.Data
            });
        }
    }

    const accToMintMap = {};
    const accToOwnerMap = {};

    if (tx.Meta && tx.Meta.InnerInstructions) {
        for (const inner of tx.Meta.InnerInstructions) {
            if (inner.Instructions) {
                for (const innerIx of inner.Instructions) {
                    instructions.push(innerIx);
                }
            }
        }
    }

    const preBalances = {};
    const postBalances = {};

    if (tx.Meta) {
        if (tx.Meta.PreTokenBalances) {
            for (const bal of tx.Meta.PreTokenBalances) {
                preBalances[bal.AccountIndex] = bal;
                if (bal.AccountIndex >= accountKeys.length) continue;
                const key = accountKeys[bal.AccountIndex];
                if (key && bal.Mint) accToMintMap[key.toBase58()] = bal.Mint;
                if (key && bal.Owner) accToOwnerMap[key.toBase58()] = bal.Owner;
            }
        }
        if (tx.Meta.PostTokenBalances) {
            for (const bal of tx.Meta.PostTokenBalances) {
                postBalances[bal.AccountIndex] = bal;
                if (bal.AccountIndex >= accountKeys.length) continue;
                const key = accountKeys[bal.AccountIndex];
                if (key && bal.Mint) accToMintMap[key.toBase58()] = bal.Mint;
                if (key && bal.Owner) accToOwnerMap[key.toBase58()] = bal.Owner;
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
