package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

type Entry struct {
	Timestamp      string `json:"timestamp"` // used as ID
	DeviceName     string `json:"deviceName"`
	Attribute      string `json:"attribute"`
	AttributeValue string `json:"attributeValue"`
}

// ============================================================================================================================
// Init - reset all the things
// Init is called during chaincode instantiation to initialize any data.
// Chaincode upgrade also calls this function to reset or to migrate data.
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return nil, nil
}

// ============================================================================================================================
// Invoke - Entry point for Invocations
// Invoke is called per transaction on the chaincode.
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "create" {
		return t.createEntry(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation: " + function)
}

// ============================================================================================================================
// Query - Entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "adHocQuery" { //find entries based on an ad hoc rich query
		return t.adHocQuery(stub, args)
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

// ============================================================================================================================
// Create Entry - create a new entry, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) createEntry(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//   0       	1       		2    		 3
	// "timestamp", "deviceName", "attribute", "attributeValue"
	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	fmt.Println("- start entry creation")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, errors.New("4th argument must be a non-empty string")
	}
	timestamp := args[0]
	deviceName := args[1]
	attribute := args[2]
	attributeValue := args[3]

	//check if entry already exists
	entryAsBytes, err := stub.GetState(timestamp)
	if err != nil {
		return nil, errors.New("Failed to get entry: " + err.Error())
	} else if entryAsBytes != nil {
		fmt.Println("This entry already exists: " + timestamp)
		return nil, errors.New("This entry already exists: " + timestamp)
	}

	// ==== Create Entry object and marshal to JSON ====
	entry := &Entry{timestamp, deviceName, attribute, attributeValue}
	entryJSONasBytes, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}

	// Save entry to state
	err = stub.PutState(timestamp, entryJSONasBytes)
	if err != nil {
		return nil, err
	}

	fmt.Println("- end entry creation")
	return nil, nil
}

// ===== Ad hoc rich query ========================================================
// This method uses a query string to perform a rich query.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// =========================================================================================
func (t *SimpleChaincode) adHocQuery(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	//   0
	// "queryString"
	if len(args) < 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return nil, err
	}
	return queryResults, nil
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {
	
	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}
