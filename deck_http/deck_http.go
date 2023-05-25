package deck_http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"tui-deck/deck_structs"
	"tui-deck/utils"
)

func httpCall(jsonBody []byte, method string, url string, user string, password string) (*http.Response, error) {
	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+basicAuth(user, password))
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return res, errors.New(res.Status)
	}
	return res, nil
}

func basicAuth(username, password string) string {
	auth := fmt.Sprintf("%s:%s", username, password)
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func GetBoards(configuration utils.Configuration) ([]deck_structs.Board, error) {
	call, err := httpCall(nil, http.MethodGet,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.0/boards", configuration.Url),
		configuration.User, configuration.Password)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(call.Body)
	var boards []deck_structs.Board
	err = decoder.Decode(&boards)
	if err != nil {
		panic(err)
	}
	return boards, nil
}

func GetStacks(boardId int, configuration utils.Configuration) ([]deck_structs.Stack, error) {
	call, err := httpCall(nil, http.MethodGet,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.0/boards/%d/stacks", configuration.Url, boardId),
		configuration.User, configuration.Password)
	if err != nil {
		return nil, err

	}
	decoder := json.NewDecoder(call.Body)
	var stacks []deck_structs.Stack

	err = decoder.Decode(&stacks)
	if err != nil {
		panic(err)
	}
	return stacks, nil
}

func UpdateCard(boardId int, stackId int, cardId int, jsonBody string, configuration utils.Configuration) (deck_structs.Card, error) {
	body := []byte(jsonBody)

	call, err := httpCall(body, http.MethodPut,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.0/boards/%d/stacks/%d/cards/%d", configuration.Url, boardId, stackId, cardId),
		configuration.User, configuration.Password)
	if err != nil {
		return deck_structs.Card{}, err

	}
	decoder := json.NewDecoder(call.Body)
	var card deck_structs.Card

	err = decoder.Decode(&card)
	if err != nil {
		panic(err)
	}
	return card, nil
}

func GetBoardDetail(boardId int, configuration utils.Configuration) (deck_structs.Board, error) {
	call, err := httpCall(nil, http.MethodGet,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.0/boards/%d", configuration.Url, boardId),
		configuration.User, configuration.Password)
	if err != nil {
		return deck_structs.Board{}, err

	}
	decoder := json.NewDecoder(call.Body)
	var board deck_structs.Board

	err = decoder.Decode(&board)
	if err != nil {
		panic(err)
	}
	return board, nil
}

func DeleteLabel(boardId int, stackId int, cardId int, jsonBody string, configuration utils.Configuration) (deck_structs.Card, error) {
	body := []byte(jsonBody)

	call, err := httpCall(body, http.MethodPut,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.0/boards/%d/stacks/%d/cards/%d/removeLabel", configuration.Url, boardId, stackId, cardId),
		configuration.User, configuration.Password)
	if err != nil {
		return deck_structs.Card{}, err

	}
	decoder := json.NewDecoder(call.Body)
	var card deck_structs.Card

	err = decoder.Decode(&card)
	if err != nil {
		panic(err)
	}
	return card, nil
}

func AssignLabel(boardId int, stackId int, cardId int, jsonBody string, configuration utils.Configuration) (deck_structs.Card, error) {
	body := []byte(jsonBody)

	call, err := httpCall(body, http.MethodPut,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.0/boards/%d/stacks/%d/cards/%d/assignLabel", configuration.Url, boardId, stackId, cardId),
		configuration.User, configuration.Password)
	if err != nil {
		return deck_structs.Card{}, err

	}
	decoder := json.NewDecoder(call.Body)
	var card deck_structs.Card

	err = decoder.Decode(&card)
	if err != nil {
		panic(err)
	}
	return card, nil
}
