package handlers

import (
	"context"
	"encoding/json"

	"git.zabbix.com/ap/plugin-support/zbxerr"
	"go.mongodb.org/mongo-driver/bson"
)

func TablespacesUsageHandler(ctx context.Context, s Database, _ map[string]string) (interface{}, error) {
	colUsage := &bson.M{}

	jsonRes, err := json.Marshal(colUsage)
	if err != nil {
		return nil, zbxerr.ErrorCannotMarshalJSON.Wrap(err)
	}

	return string(jsonRes), nil
}
