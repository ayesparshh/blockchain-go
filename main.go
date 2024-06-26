package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Block struct{
	Pos          	int   
	Data			Minecheckout		
	Timestamp		string
	Hash 			string
	Prevhash		string

}


type Mine struct {

   ID 			string  `json:"id"`
   Title 		string	`json:"title"`
   Miner 		string	`json:"miner"`
   Miningdate   string	`json:"miningdate"`
   ISBN 		string	`json:"isbn"`
}

type Minecheckout struct {
	MineID 			string	`json:"mineid"`
	Miner 			string	`json:"miner"`
	Checkoutdate 	string	`json:"checkoutdate"`
	IsGeneis 		bool	`json:"isgeneis"`
}

type Blockchain struct {
	Blocks []*Block
}

var blockchain *Blockchain

func (b *Block) generatehash() {
	bytes , _ := json.Marshal(b.Data)

	data := string(b.Pos) + string(bytes) + b.Prevhash + b.Timestamp

	hash := sha256.New()

	hash.Write([]byte(data))

	b.Hash = hex.EncodeToString(hash.Sum(nil)) 

}	

func CreateBlock(prevBlock *Block, data Minecheckout) *Block {

	block := &Block{}

	block.Pos = prevBlock.Pos + 1
	block.Timestamp = time.Now().String()
	block.Prevhash  = prevBlock.Hash
	block.generatehash()

	return block	
}	

func (b *Block)validateHash(hash string) bool {
	b.generatehash()
	
	return b.Hash == hash
}

func validBlock(block , prevBlock *Block) bool {

	if prevBlock.Hash != block.Prevhash {
		return false
	}

	if !block.validateHash(block.Hash) {
		return false
	}

	if prevBlock.Pos + 1 != block.Pos {
		return false
	}

	return true
}

func (b *Blockchain) AddBlock(data Minecheckout) {
	prevBlock := b.Blocks[len(b.Blocks)-1]

	block := CreateBlock(prevBlock, data)

	if validBlock(block, prevBlock) {
		b.Blocks = append(b.Blocks, block)
	}
}	

func writeBlock(w http.ResponseWriter, r *http.Request) {
	var chechout Minecheckout
	if err := json.NewDecoder(r.Body).Decode(&chechout); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		w.Write([]byte("could not write block"))
		return
	}
	blockchain.AddBlock(chechout)

	res,err := json.MarshalIndent(chechout, "", "  ")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not marshal mine %v", err)
		w.Write([]byte("could not save new mine data"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func newMine(w http.ResponseWriter, r *http.Request) {
	var mine Mine
	if err := json.NewDecoder(r.Body).Decode(&mine); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		w.Write([]byte("could not create new mine"))
		return
	}

	h := md5.New()

	io.WriteString(h, mine.ISBN+mine.Miningdate)
	mine.ID = fmt.Sprintf("%x", h.Sum(nil))

	resp, err := json.MarshalIndent(mine, "", "  ")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not marshal mine %v", err)
		w.Write([]byte("could not save new mine data"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)

}	

func GenesisBlock() *Block {
	return CreateBlock(&Block{}, Minecheckout{IsGeneis: true})
}

func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{GenesisBlock()}}
}

func getBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes,err := json.MarshalIndent(blockchain.Blocks, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		w.Write([]byte("could not get blockchain"))
		return
	}

	io.WriteString(w, string(bytes))
}


func main () {

	blockchain  =  NewBlockchain()

	r := mux.NewRouter()
	r.HandleFunc("/", getBlockchain).Methods("GET")
	r.HandleFunc("/", writeBlock).Methods("POST")
	r.HandleFunc("/new", newMine).Methods("POST") 

	go func() {
		for _ , block := range blockchain.Blocks {
			fmt.Printf("Prevhash: %x\n", block.Prevhash)
			bytes, _ :=json.MarshalIndent(block.Data, "", "  ")
			fmt.Printf("Data: %s\n", string(bytes))
			fmt.Printf("Timestamp: %s\n", block.Timestamp)
			fmt.Printf("Hash: %x\n", block.Hash)
		}
	}()		
	log.Println("listening on port 3000")

	log.Fatal(http.ListenAndServe(":3000",r))
}