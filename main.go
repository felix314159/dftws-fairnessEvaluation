// This program is used to evaluate the fairness of DFTWS winner selection (https://arxiv.org/abs/2312.01951)
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	amountOfRunTrials = 10 // i define a run trial as a set of node identities that are tested for winner selection in many different runs
	amountOfRunsPerIdentitySet = 3000
	amountOfNodeIdentitiesPerRun = 40
	csvOutputName = "data.csv"
)

func main() {


	fmt.Printf("Simulation Config:\n\tAmount of trials: %v\n\tAmount of runs per trial: %v\n\tAmount of miner identities per run: %v\n\n", amountOfRunTrials, amountOfRunsPerIdentitySet, amountOfNodeIdentitiesPerRun)

	// ----

	for j:=0; j<amountOfRunTrials; j++ {
		fmt.Printf("---- Run Trial %v ----\n", j+1) // start counting at 1, its more intuitive when reading output

		// generate miner identities
		minerSlice := []Miner{}
		for k:=0; k<amountOfNodeIdentitiesPerRun; k++ {
			curMiner := NewMiner()
			minerSlice = append(minerSlice, curMiner)
		}

		// print which miner identities have been generated
		for ind, mTemp := range minerSlice {
			fmt.Printf("\tNode %v: %v\n", ind+1, mTemp.NodeID) // start counting at 1, its more intuitive when reading output
		}
		
		// keep track of how often each miner won (for each run different values for raCommit and solutionHash are simulated)
		leaderboard := make(map[string]int)

		for i:=0; i<amountOfRunsPerIdentitySet; i++ {
			winnerNodeID := RunWinnerSelectionWrapper(minerSlice, i)
			
			// keep track of winner
			 if _, exists := leaderboard[winnerNodeID]; exists {
	            // node won already before, so increase its win counter by 1
	            leaderboard[winnerNodeID]++
	        } else {
	            //  node won for the first time, so set its win counter to 1
	            leaderboard[winnerNodeID] = 1
	        }

		}

		// print winner stats for all runs that were performed here
		for ind, z := range minerSlice {
			fmt.Printf("\tAmount of node %v wins: %v\n", ind+1, leaderboard[z.NodeID]) // start counting at 1, its more intuitive when reading output
		}
		
		// write the data to csv so that it can be analyzed later
		// 		if the csv file already exists at first iteration, delete it and then create a new one and write its header
		if j == 0 {
			// delete old csv if it exists
			DeleteExistingCSV(csvOutputName)

			// create new csv with n1,n2,... header
			CreateNewCSVWithHeader(csvOutputName)
		}

		// 		define row to write to csv
		var row []string
		for _, m := range minerSlice {
			// convert int to string and append it to row
			row = append(row, fmt.Sprintf("%v", leaderboard[m.NodeID]))
		}
		// write row to csv
		AppendRowToCSV(row, csvOutputName)

	}

}

// RunWinnerSelectionWrapper takes a slice of miner identities and a run index for a better overview, and then simulates everything involved in winner selection with pseudo random values for raCommitSecret and solutionHash. Returns the winner nodeID string.
func RunWinnerSelectionWrapper(minerSlice []Miner, runIndex int) string {
	// generate imaginary solution hash
	pseudoRandomBytes := make([]byte, 16)
	_, err := rand.Read(pseudoRandomBytes)
	if err != nil {
		panic(err)
	}
	pseudoRandomString := fmt.Sprintf("%x", pseudoRandomBytes)
	solutionHash := NewHash(pseudoRandomString)

	// simulate how the miners would sign such a hash and get their commitments
	eligibleMiners := []ActiveMiner{}
	for _, m := range minerSlice {
		mCommitment := GetMinerCommitment(solutionHash, m) // returns ActiveMiner which contains the commitment
		eligibleMiners = append(eligibleMiners, mCommitment)
	}
	
	// ---- winner selection ----

	// 			generate raCommitSecret
	secretBytes := make([]byte, 71)
	_, err = rand.Read(secretBytes)
	if err != nil {
		panic(err)
	}
	raCommitSecret := fmt.Sprintf("%x", secretBytes) // convert to hex string of length 2*71 = 142

	// 			run winner selection algo to determine which miner would be chosen winner in this scenario
	
	blockWinnerNodeID := WinnerSelection(eligibleMiners, raCommitSecret)

	// fmt.Printf("Winner of run %v: %v\n", runIndex+1, blockWinnerNodeID) // start counting at 1, its more intuitive when reading output
	
	return blockWinnerNodeID
}

func WinnerSelection(inputEligibleMiners []ActiveMiner, raCommitSecret string) string {
	if raCommitSecret == "" {
		panic("WinnerSelection - My raCommitSecret is still an empty string! This means I never got the message so I can not perform the winner selection. Will panic now, sorry!")
	}

	// store concated sigs
	c := ""

	//	0. sort list by nodeID ascending
	//			0.1 extract nodeID strings
	var nodeIDStrings []string
	for _, m := range inputEligibleMiners {
		nodeIDStrings = append(nodeIDStrings, m.Commitment.OriginalSenderNodeID)
	}

	//			0.2 sort string slice in place (ascending), 0<9<a<z and case-insensitive
	sort.Slice(nodeIDStrings, func(i, j int) bool {
		return strings.ToLower(nodeIDStrings[i]) < strings.ToLower(nodeIDStrings[j])
	})

	//			0.3 re-create active miner slice in the correct order + concatenate sigs in this order
	var eligibleMinersSorted []ActiveMiner
	for _, n := range nodeIDStrings {
		for _, o := range inputEligibleMiners {
			if n == o.Commitment.OriginalSenderNodeID {
				eligibleMinersSorted = append(eligibleMinersSorted, o)
				c += string(o.Commitment.SigCommit)
			}
		}
	}

	//  1. sigs have already been concatenated
	//	2. take secretCommitBytes (revealed by RA) and calculate Keccak256(c + raCommitSecret)
	a := NewHash(c + raCommitSecret).GetString()

	// 3. take first 15 chars of a and convert to decimal number
	aDec, err := strconv.ParseInt(a[:15], 16, 64)
	if err != nil {
		panic(err)
	}

	//	4. calculate winner index: hash_to_number_result % len(activeMinerList)
	winnerIndex := aDec % int64(len(eligibleMinersSorted))

	//	5. access winner with index
	winnerNodeAddress := eligibleMinersSorted[winnerIndex].Commitment.OriginalSenderNodeID

	return winnerNodeAddress
}

// ---- Hash ----

// Hash has only one field which is the hash as slice of byte.
type Hash struct {
	Bytes []byte
}

// NewHash is the constructor of Hash which makes sure that the hash function is used correctly.
func NewHash(input string) Hash {
	h := sha3.NewLegacyKeccak256() 			// create hash.Hash object [i used legacy for compatibility with online comparisions like https://emn178.github.io/online-tools/keccak_256.html]
	h.Write([]byte(input))          		// fill object with data
	var hashResult Hash = Hash{h.Sum(nil)}	// calculate hashsum (nil means store in new byte slice)
	return hashResult
}

// GetString returns the hash as string.
func (h Hash) GetString() string {
	if h.Bytes == nil {
		panic("hash.Hash GetString - The Bytes field of this hash.Hash is nil! Aborting..")
	}
	return hex.EncodeToString(h.Bytes)
}

// ---- Active Miner ----

// Miner was created to make this demonstration for flexible with regards to easily being able to re-use an identity for many runs
type Miner struct {
	NodeID  string
	Pub     crypto.PubKey
	Priv 	crypto.PrivKey
}

// NewMiner is the constructor used to generate a miner instance
func NewMiner() Miner {
	// create first node identity
	_, ed25519Priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}

	// 		convert ed25519 priv to libp2p priv
	priv, err := crypto.UnmarshalEd25519PrivateKey(ed25519Priv)
    if err != nil {
        panic(err)
    }

    // 		derive libp2p pub
    pub := priv.GetPublic()

	// 		derive libp2p nodeID from public key
	MyNodeIDString, err := PubKeyToNodeID(pub)
	if err != nil {
		panic(err)
	}

	return Miner {
		NodeID: MyNodeIDString,
		Pub: pub,
		Priv: priv,
	}
}

// ActiveMiner is used to aggregate information about an active miner that has broadcast commitments before problem expiry.
type ActiveMiner struct {
	Commitment 					MinerCommitment 	// Commitment 1: Hash(solutionHash) that this miner broadcast, Commitment 2: Signature Sig(solutionHash) that this miner broadcast [solutionHash unknown at least until problem expires]
}

// NewActiveMiner is the constructor for ActiveMiner.
func NewActiveMiner(commitment MinerCommitment) ActiveMiner {
	return ActiveMiner {
		Commitment: commitment,
	}
}

// ---- MinerCommitment ----

// MinerCommitment is a struct used after a miner solved the block problem to store original sender ID, Hash(Hash(solutiondata)) and Sig(solutionHash)
type MinerCommitment struct {
	OriginalSenderNodeID  	string 		// NodeID who this commitment is originally from. This is required because pubsub.Message.ReceivedFrom is the address of whoever forwarded the message to you, but this does not have to be the original sender!
	HashCommit				Hash 		// Hash of hash (prove you know the hash without revealing it)
	SigCommit 				[]byte 		// Sig of hash (so that miners can not just re-broadcast the Hash-of-hashes of other miners' solutions)
}

// ---- Libp2p node identity ----

// PubKeyToNodeID takes a PubKey and returns the human-readable node ID (12D3Koo...)
func PubKeyToNodeID(pubKeyObject crypto.PubKey) (string, error) {
	peerID, err := peer.IDFromPublicKey(pubKeyObject)
	if err != nil {
		return "", fmt.Errorf("Failed to convert PubKey to peerID: %v", err)
	}

	return peerID.String(), nil
}

// GetMinerCommitment takes an imaginary solution hash and a libp2p node identity, then it determines how that node would have signed its result and returns the resulting commitment wrapped in an ActiveMiner instance
func GetMinerCommitment(solutionHash Hash, miner Miner) ActiveMiner {
	// derive hash commits of nodes that would have gotten that result, H(H(solutionData)) = H(solutionHashString)
	hashCommit := NewHash(solutionHash.GetString())

	// 	sign imaginary solution hash
	solSig, err := miner.Priv.Sign(solutionHash.Bytes)
	if err != nil {
		panic(err)
	}

	commit := MinerCommitment {
		OriginalSenderNodeID: miner.NodeID,
		HashCommit: hashCommit, 
		SigCommit: solSig, 
	}

	return NewActiveMiner(commit)
}

// ---- csv ----

// DeleteExistingCSV deletes the csv file if it already exists.
func DeleteExistingCSV(csvFileLocation string){
	// check if file exists
	_, err := os.Stat(csvFileLocation)
	if err == nil {
		// file does exist, so delete it
		err = os.Remove(csvFileLocation)
		if err != nil {
			panic(err)
		}
	}

}

// CreateNewCSVWithHeader creates a new csv with the n1,n2,... header.
func CreateNewCSVWithHeader(csvFileLocation string) {
	// create new csv
	file, err := os.Create(csvFileLocation)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// define header (n1, n2, ...)
	var header []string
	for temp:=0; temp<amountOfNodeIdentitiesPerRun; temp++ {
		header = append(header, fmt.Sprintf("n%v", temp+1))
	}

	// write header to file
	AppendRowToCSV(header, csvFileLocation)

}

// AppendRowToCSV appends the given row to the csv.
func AppendRowToCSV(row []string, csvFileLocation string) {
	// open csv
	file, err := os.OpenFile(csvFileLocation, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// write data
	if err := writer.Write(row); err != nil {
		panic(err)
	}
}
