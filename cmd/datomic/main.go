package main

import (
	"encoding/json"
	"errors"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Transaction struct {
	Op     string
	Values []int
	Key    int
}

type TransactionRequest struct {
	maelstrom.MessageBody
	Transactions [][]any `json:"txn"`
}

func deserializeTrasactions(raw [][]any) ([]*Transaction, error) { //nolint: gocognit
	transactions := []*Transaction{}
	for _, t := range raw {
		op, ok := t[0].(string)
		if !ok {
			return nil, errors.New("bad transaction format: op")
		}

		key, ok := t[1].(float64)
		if !ok {
			return nil, errors.New("bad transaction format: key")
		}

		values := []int{}
		if t[2] != nil { //nolint: nestif
			rawValues, isList := t[2].([]any)
			rawValue, isNumber := t[2].(float64)
			if !isList && !isNumber {
				return nil, errors.New("bad transaction format: values")
			}

			if isNumber {
				values = append(values, int(rawValue))
			}

			if isList {
				for _, v := range rawValues {
					numValue, ok := v.(float64)
					if !ok {
						return nil, errors.New("bad transaction format: values")
					}

					values = append(values, int(numValue))
				}
			}
		}

		transactions = append(transactions, &Transaction{
			Op:     op,
			Key:    int(key),
			Values: values,
		})
	}

	return transactions, nil
}

func serializeTransactions(transactions []*Transaction) [][]any {
	result := [][]any{}

	for _, t := range transactions {
		var resultValue any
		if t.Op == "append" {
			resultValue = t.Values[0]
		} else {
			resultValue = t.Values
		}

		result = append(result, []any{
			t.Op,
			t.Key,
			resultValue,
		})
	}

	return result
}

func main() {
	n := maelstrom.NewNode()

	state := NewState()

	n.Handle("txn", func(msg maelstrom.Message) error {
		var body TransactionRequest
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		transactions, err := deserializeTrasactions(body.Transactions)
		if err != nil {
			return err
		}

		state.Transact(transactions)

		result := serializeTransactions(transactions)

		return n.Reply(msg, map[string]any{
			"type": "txn_ok",
			"txn":  result,
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
