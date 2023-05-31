package deck_db

import (
	"encoding/json"
	"fmt"
	"os"
	"tui-deck/deck_http"
	"tui-deck/deck_structs"
	"tui-deck/utils"
)

func GetBoardDetails(boardId int, updateBoard bool, configuration utils.Configuration) (deck_structs.Board, error) {
	currentBoard := deck_structs.Board{}
	var err error
	if updateBoard {
		currentBoard, err = deck_http.GetBoardDetail(boardId, configuration)
		if err != nil {
			return deck_structs.Board{}, err
		}
		var boardDetailFile *os.File
		boardDetailFile, err = utils.CreateFile(fmt.Sprintf("%s/db/board-detail-%d.json", configuration.ConfigDir, boardId))
		if err != nil {
			return deck_structs.Board{}, err
		}
		var marshal []byte
		marshal, err = json.Marshal(currentBoard)
		if err != nil {
			return deck_structs.Board{}, err
		}
		_, err = boardDetailFile.Write(marshal)
		if err != nil {
			return deck_structs.Board{}, err
		}
	} else {
		var localBoardFile *os.File
		localBoardFile, err = os.Open(fmt.Sprintf("%s/db/board-%d.json", configuration.ConfigDir, boardId))
		if err != nil {
			return deck_structs.Board{}, err
		}
		decoder := json.NewDecoder(localBoardFile)
		err = decoder.Decode(&currentBoard)
		if err != nil {
			return deck_structs.Board{}, err
		}
	}
	currentBoard.Updated = updateBoard
	return currentBoard, nil
}

func GetStacks(boardId int, updateBoard bool, configuration utils.Configuration) ([]deck_structs.Stack, error) {
	stacks := make([]deck_structs.Stack, 0)
	var err error
	if updateBoard {
		stacks, err = deck_http.GetStacks(boardId, configuration)
		if err != nil {
			return nil, err
		}
		var stacksFile *os.File
		stacksFile, err = utils.CreateFile(fmt.Sprintf("%s/db/stacks-%d.json", configuration.ConfigDir, boardId))
		if err != nil {
			return nil, err
		}
		var marshal []byte
		marshal, err = json.Marshal(stacks)
		if err != nil {
			return nil, err
		}
		_, err = stacksFile.Write(marshal)
		if err != nil {
			return nil, err
		}
	} else {
		var localStacks *os.File
		localStacks, err = os.Open(fmt.Sprintf("%s/db/stacks-%d.json", configuration.ConfigDir, boardId))
		if err != nil {
			return nil, err
		}
		decoder := json.NewDecoder(localStacks)
		err = decoder.Decode(&stacks)
		if err != nil {
			return nil, err
		}
	}
	return stacks, nil
}
