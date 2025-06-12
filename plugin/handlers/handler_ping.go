package handlers

import "context"

func PingHandler(ctx context.Context, s Database, _ map[string]string) (interface{}, error) {
	/*if err := s.Database.Ping(); err != nil {
		Logger.Debugf("ping failed, %s", err.Error())

		return PingFailed, nil
	}*/

	return PingOk, nil
}
