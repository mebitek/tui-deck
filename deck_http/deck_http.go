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

func httpCall(jsonBody []byte, method string, url string, user string, password string, ocs bool) (*http.Response, error) {
	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Add("Authorization", "Basic "+basicAuth(user, password))
	if ocs {
		req.Header.Add("OCS-APIRequest", "true")
	}
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
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards", configuration.Url),
		configuration.User, configuration.Password, false)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(call.Body)
	var boards []deck_structs.Board
	err = decoder.Decode(&boards)
	if err != nil {
		panic(err)
	}

	filteredBoards := make([]deck_structs.Board, 0)
	for _, b := range boards {
		if b.DeletedAt == 0 {
			filteredBoards = append(filteredBoards, b)
		}
	}

	return filteredBoards, nil
}

func GetStacks(boardId int, configuration utils.Configuration) ([]deck_structs.Stack, error) {
	call, err := httpCall(nil, http.MethodGet,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards/%d/stacks", configuration.Url, boardId),
		configuration.User, configuration.Password, false)
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

func AddCard(boardId int, stackId int, jsonBody string, configuration utils.Configuration) (deck_structs.Card, error) {
	body := []byte(jsonBody)

	call, err := httpCall(body, http.MethodPost,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards/%d/stacks/%d/cards", configuration.Url, boardId, stackId),
		configuration.User, configuration.Password, false)
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

func UpdateCard(boardId int, stackId int, cardId int, jsonBody string, configuration utils.Configuration) (deck_structs.Card, error) {
	body := []byte(jsonBody)

	call, err := httpCall(body, http.MethodPut,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards/%d/stacks/%d/cards/%d", configuration.Url, boardId, stackId, cardId),
		configuration.User, configuration.Password, false)
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
func DeleteCard(boardId int, stackId int, cardId int, configuration utils.Configuration) (deck_structs.Card, error) {

	call, err := httpCall(nil, http.MethodDelete,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards/%d/stacks/%d/cards/%d", configuration.Url, boardId, stackId, cardId),
		configuration.User, configuration.Password, false)
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
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards/%d", configuration.Url, boardId),
		configuration.User, configuration.Password, false)
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
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards/%d/stacks/%d/cards/%d/removeLabel", configuration.Url, boardId, stackId, cardId),
		configuration.User, configuration.Password, false)
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
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards/%d/stacks/%d/cards/%d/assignLabel", configuration.Url, boardId, stackId, cardId),
		configuration.User, configuration.Password, false)
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

func AddBoard(jsonBody string, configuration utils.Configuration) (deck_structs.Board, error) {
	body := []byte(jsonBody)

	call, err := httpCall(body, http.MethodPost,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards", configuration.Url),
		configuration.User, configuration.Password, false)
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

func EditBoard(boardId int, jsonBody string, configuration utils.Configuration) (deck_structs.Board, error) {
	body := []byte(jsonBody)

	call, err := httpCall(body, http.MethodPut,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards/%d", configuration.Url, boardId),
		configuration.User, configuration.Password, false)
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

func DeleteBoard(boardId int, configuration utils.Configuration) (deck_structs.Board, error) {

	call, err := httpCall(nil, http.MethodDelete,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards/%d", configuration.Url, boardId),
		configuration.User, configuration.Password, false)
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

func DeleteBoardLabel(boardId int, labelId int, configuration utils.Configuration) (int, error) {

	call, err := httpCall(nil, http.MethodDelete,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards/%d/labels/%d", configuration.Url, boardId, labelId),
		configuration.User, configuration.Password, false)
	if err != nil {
		return call.StatusCode, err
	}
	return call.StatusCode, nil
}

func AddBoardLabel(boardId int, jsonBody string, configuration utils.Configuration) (deck_structs.Label, error) {
	body := []byte(jsonBody)

	call, err := httpCall(body, http.MethodPost,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards/%d/labels", configuration.Url, boardId),
		configuration.User, configuration.Password, false)
	if err != nil {
		return deck_structs.Label{}, err

	}
	decoder := json.NewDecoder(call.Body)
	var label deck_structs.Label

	err = decoder.Decode(&label)
	if err != nil {
		panic(err)
	}
	return label, nil
}

func EditBoardLabel(boardId int, labelId int, jsonBody string, configuration utils.Configuration) (deck_structs.Label, error) {
	body := []byte(jsonBody)

	call, err := httpCall(body, http.MethodPut,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards/%d/labels/%d", configuration.Url, boardId, labelId),
		configuration.User, configuration.Password, false)
	if err != nil {
		return deck_structs.Label{}, err

	}
	decoder := json.NewDecoder(call.Body)
	var label deck_structs.Label

	err = decoder.Decode(&label)
	if err != nil {
		panic(err)
	}
	return label, nil
}

func AddStack(boardId int, jsonBody string, configuration utils.Configuration) (deck_structs.Stack, error) {
	body := []byte(jsonBody)

	call, err := httpCall(body, http.MethodPost,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards/%d/stacks", configuration.Url, boardId),
		configuration.User, configuration.Password, false)
	if err != nil {
		return deck_structs.Stack{}, err

	}
	decoder := json.NewDecoder(call.Body)
	var stack deck_structs.Stack

	err = decoder.Decode(&stack)
	if err != nil {
		panic(err)
	}
	return stack, nil
}

func DeleteStack(boardId int, stackId int, configuration utils.Configuration) (int, error) {

	call, err := httpCall(nil, http.MethodDelete,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards/%d/stacks/%d", configuration.Url, boardId, stackId),
		configuration.User, configuration.Password, false)
	if err != nil {
		return call.StatusCode, err
	}
	return call.StatusCode, nil
}

func EditStack(boardId int, stackId int, jsonBody string, configuration utils.Configuration) (deck_structs.Stack, error) {
	body := []byte(jsonBody)

	call, err := httpCall(body, http.MethodPut,
		fmt.Sprintf("%s/index.php/apps/deck/api/v1.1/boards/%d/stacks/%d", configuration.Url, boardId, stackId),
		configuration.User, configuration.Password, false)
	if err != nil {
		return deck_structs.Stack{}, err

	}
	decoder := json.NewDecoder(call.Body)
	var stack deck_structs.Stack

	err = decoder.Decode(&stack)
	if err != nil {
		panic(err)
	}
	return stack, nil
}

func GetComments(cardId int, configuration utils.Configuration) ([]deck_structs.Comment, error) {

	call, err := httpCall(nil, http.MethodGet,
		fmt.Sprintf("%s/ocs/v2.php/apps/deck/api/v1.0/cards/%d/comments", configuration.Url, cardId),
		configuration.User, configuration.Password, true)

	if err != nil {
		return nil, err

	}

	var ocs deck_structs.OcsResponse

	decoder := json.NewDecoder(call.Body)

	err = decoder.Decode(&ocs)
	if err != nil {
		panic(err)
	}
	return ocs.Ocs.Data, nil
}

func AddComment(cardId int, jsonBody string, configuration utils.Configuration) (deck_structs.Comment, error) {
	body := []byte(jsonBody)

	call, err := httpCall(body, http.MethodPost,
		fmt.Sprintf("%s/ocs/v2.php/apps/deck/api/v1.0/cards/%d/comments", configuration.Url, cardId),
		configuration.User, configuration.Password, true)
	if err != nil {
		return deck_structs.Comment{}, err

	}
	decoder := json.NewDecoder(call.Body)
	var ocs deck_structs.OcsResponseSingle
	err = decoder.Decode(&ocs)
	if err != nil {
		panic(err)
	}
	return ocs.Ocs.Data, nil
}

func EditComment(cardId int, commentid int, jsonBody string, configuration utils.Configuration) (deck_structs.Comment, error) {
	body := []byte(jsonBody)

	call, err := httpCall(body, http.MethodPut,
		fmt.Sprintf("%s/ocs/v2.php/apps/deck/api/v1.0/cards/%d/comments/%d", configuration.Url, cardId, commentid),
		configuration.User, configuration.Password, true)
	if err != nil {
		return deck_structs.Comment{}, err

	}
	decoder := json.NewDecoder(call.Body)
	var ocs deck_structs.OcsResponseSingle
	err = decoder.Decode(&ocs)
	if err != nil {
		panic(err)
	}
	return ocs.Ocs.Data, nil
}

func DeleteComment(cardId int, commentId int, configuration utils.Configuration) (int, error) {
	call, err := httpCall(nil, http.MethodDelete,
		fmt.Sprintf("%s/ocs/v2.php/apps/deck/api/v1.0/cards/%d/comments/%d", configuration.Url, cardId, commentId),
		configuration.User, configuration.Password, true)
	if err != nil {
		return call.StatusCode, err
	}
	return call.StatusCode, nil
}
