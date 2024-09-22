package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	// "strconv"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"

	// banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/CoreumFoundation/coreum/v4/pkg/client"
	coreumconfig "github.com/CoreumFoundation/coreum/v4/pkg/config"
	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	"github.com/rs/cors"
)

const (
	// Replace it with your own mnemonic
	senderMnemonic    = ""
	recipientMnemonic = ""

	chainID       = constant.ChainIDTest
	addressPrefix = constant.AddressPrefixTest
	nodeAddress   = "full-node.testnet-1.coreum.dev:9090"
)

type Response struct {
    Message       string `json:"message"`
    TransactionID string `json:"transaction_id"`
    Denom         string `json:"denom,omitempty"`
    IssuerAddress string `json:"issuer_address,omitempty"`
}

//  struct is serialized to JSON or deserialized from JSON,

type IssueTokenRequest struct {
	Symbol        string `json:"symbol"`
	Subunit       string `json:"subunit"`
	Precision     int    `json:"precision"`
	InitialAmount string `json:"initial_amount"`
	Description   string `json:"description"`
}

type TransferAdminRequest struct {
	Denom    string `json:"denom"`
	NewAdmin string `json:"new_admin"`
}

type ClearAdminRequest struct {
	Denom string `json:"denom"`
}

func main() {

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(addressPrefix, addressPrefix+"pub")
	config.SetCoinType(constant.CoinType)
	config.Seal()

	r := mux.NewRouter()

	r.HandleFunc("/api/issue-token", issueTokenHandler).Methods("POST")
	r.HandleFunc("/api/transfer-admin", transferAdminHandler).Methods("POST")
	r.HandleFunc("/api/clear-admin", clearAdminHandler).Methods("POST")

	handler := cors.Default().Handler(r)

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))

}

func issueTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req IssueTokenRequest
	// When a client (like a frontend app or API consumer) sends a request, 
	// the data is often included in the request body in JSON format
	// We decode the request body so that we can extract the data sent 
	// by the client in a structured way and use it in the server-side logic.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Context is used in Go to manage deadlines, 
	// cancel signals, and request-scoped values

	ctx := context.Background()
	clientCtx, txFactory, senderAddress, err := setupClientContext()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error setting up client context:", err)
		return
	}

	amount, ok := sdkmath.NewIntFromString(req.InitialAmount)
	if !ok {
		// Handle the error, e.g., return a bad request response
		http.Error(w, "Invalid initial amount", http.StatusBadRequest)
		return
	}

	msgIssue := &assetfttypes.MsgIssue{
		Issuer:        senderAddress.String(),
		Symbol:        req.Symbol,
		Subunit:       req.Subunit,
		Precision:     uint32(req.Precision),
		InitialAmount: amount,
		Description:   req.Description,
		Features:      []assetfttypes.Feature{assetfttypes.Feature_freezing},
	}
	// The context is passed as the first argument, allowing for things like 
	//timeouts or cancellations during the transaction broadcast.
	txResponse, err := client.BroadcastTx(ctx, clientCtx.WithFromAddress(senderAddress), txFactory, msgIssue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error broadcasting transaction:", err)
		return
	}

	denom := req.Subunit + "-" + senderAddress.String()
	log.Printf("New token issued with denom: %s", denom)

	// Converts a Go struct into JSON to send it back as the HTTP response

	json.NewEncoder(w).Encode(Response{
		Message:       "Fungible token class issued successfully",
		TransactionID: txResponse.TxHash,
		Denom:         denom, // Add this field to your Response struct
	})
}


func transferAdminHandler(w http.ResponseWriter, r *http.Request) {
	var req TransferAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	clientCtx, txFactory, senderAddress, err := setupClientContext()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error setting up client context:", err)
		return
	}

	msgTransferAdmin := &assetfttypes.MsgTransferAdmin{
		Sender:  senderAddress.String(),
		Account: req.NewAdmin,
		Denom:   req.Denom,
	}

	txResponse, err := client.BroadcastTx(ctx, clientCtx.WithFromAddress(senderAddress), txFactory, msgTransferAdmin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error broadcasting transaction:", err)
		return
	}

	json.NewEncoder(w).Encode(Response{
		Message:       "Admin rights transferred successfully",
		TransactionID: txResponse.TxHash,
	})
}

func clearAdminHandler(w http.ResponseWriter, r *http.Request) {
	var req ClearAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println("Error decoding request:", err)
		return
	}

	log.Printf("Attempting to clear admin for denom: %s", req.Denom)

	ctx := context.Background()
	clientCtx, txFactory, _, err := setupClientContext()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error setting up client context:", err)
		return
	}

	// Use recipient mnemonic to generate the address
	recipientInfo, err := clientCtx.Keyring().NewAccount(
		"recipient-key-name",
		recipientMnemonic,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	if err != nil {
		http.Error(w, "Error creating recipient account", http.StatusInternalServerError)
		log.Println("Error creating recipient account:", err)
		return
	}

	recipientAddress, err := recipientInfo.GetAddress()
	if err != nil {
		http.Error(w, "Error getting recipient address", http.StatusInternalServerError)
		log.Println("Error getting recipient address:", err)
		return
	}

	log.Printf("Clearing admin using address: %s", recipientAddress.String())

	msgClearAdmin := &assetfttypes.MsgClearAdmin{
		Sender: recipientAddress.String(),
		Denom:  req.Denom,
	}

	txResponse, err := client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(recipientAddress),
		txFactory,
		msgClearAdmin,
	)
	if err != nil {
		log.Printf("Error broadcasting transaction: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Admin rights cleared successfully. TxHash: %s", txResponse.TxHash)

	json.NewEncoder(w).Encode(Response{
		Message:       "Admin rights cleared successfully",
		TransactionID: txResponse.TxHash,
	})
}

func setupClientContext() (client.Context, client.Factory, sdk.AccAddress, error) {
	modules := module.NewBasicManager(
		auth.AppModuleBasic{},
	)

	grpcClient, err := grpc.Dial(
		nodeAddress,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})),
	)
	if err != nil {
		return client.Context{}, client.Factory{}, nil, err
	}

	encodingConfig := coreumconfig.NewEncodingConfig(modules)

	clientCtx := client.NewContext(client.DefaultContextConfig(), modules).
		WithChainID(string(chainID)).
		WithGRPCClient(grpcClient).
		WithKeyring(keyring.NewInMemory(encodingConfig.Codec)).
		WithBroadcastMode(flags.BroadcastSync).
		WithAwaitTx(true)

	txFactory := client.Factory{}.
		WithKeybase(clientCtx.Keyring()).
		WithChainID(clientCtx.ChainID()).
		WithTxConfig(clientCtx.TxConfig()).
		WithSimulateAndExecute(true)

	senderInfo, err := clientCtx.Keyring().NewAccount(
		"key-name",
		senderMnemonic,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	if err != nil {
		return client.Context{}, client.Factory{}, nil, err
	}

	senderAddress, err := senderInfo.GetAddress()
	if err != nil {
		return client.Context{}, client.Factory{}, nil, err
	}

	return clientCtx, txFactory, senderAddress, nil
}
