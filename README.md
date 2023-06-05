# tui-deck
A TUI frontend for Nextcloud Deck app written in GO using the [Rich Interactive Widgets for Terminal UIs
](https://github.com/rivo/tview)

![image](https://github.com/mebitek/tui-deck/assets/1067967/4c1913be-09d0-4fea-bc67-19da89e2e9aa)

___

# features

* switch between boards
* list cards
* edit card description
* move cards between stacks
* add/remove labels from cards
* add/edit/remove labels from labels
* basic markdown viewer
* theming

### markdown features
* headings
* task list
* unordered list
* blockquotes
* code block
* bold
* italic
* bold + italic
* inline code 

# planned features

- [x] add/remove cards
- [x] add/edit/delete stacks
- [x] add/edit/delete boards
- [x] add/edit/delete boards label
- [ ] manage comments
- [ ] manage attachments
- [x] improve boot time with local storage data

# configuration

on first start, the application will create a default config.json file in $HOME/.config/tui-deck directory

```
{
  "username": "",
  "password": "",
  "url": "https://nextcloud.example.com",
  "color": "#BF40BF"
}
```

# shortcuts

 * main

    | function    | key                         |
    |-------------|-----------------------------|
    | TAB         | swtich stacks               |
    | down arrow  | move down                   |
    | up arrow    | move up                     |
    | right arrow | move card to next stack     |
    | left arrow  | move card to previous stack |
    | ENTER       | select card                 |
    | s           | switch board                |
    | r           | reload board                |
    | a           | add card                    |
    | d           | delete card                 |
    | ctrl+a      | add stack                   |
    | ctrl+e      | edit stack                  |
    | ctrl+d      | delete stack                |
    | q           | quit app                    |
    | ?           | help                        |

* view card

    | function | key                   |
    |----------|-----------------------|
    | e        | edit card description |
    | t        | edit card labels      |
    | ESC      | back to main view     |

*  edit card

    | function | key               |
    |----------|-------------------|
    | F2       | save card         |
    | ESC      | back to view card |

* edit card labels

    | function   | key                                                                                              |
    |------------|--------------------------------------------------------------------------------------------------|
    | up arrow   | move up                                                                                          |
    | down arrow | move down                                                                                        |
    | TAB        | switch between card labels and available board labels lists                                      |
    | ENTER      | if card label has been selected, delete it. if available label has been selected, add it to card |
    | ESC        | back to view card                                                                                |

* switch boards

    | function   | key               |
    |------------|-------------------|
    | up arrow   | move up           |
    | down arrow | move down         |
    | ENTER      | select board      |
    | a          | add board         |
    | e          | edit board        |
    | d          | delete board      |
    | t          | edit board labels |
    | ESC        | back to main view |

* edit board labels

    | function   | key                   |
    |------------|-----------------------|
    | up arrow   | move up               |
    | down arrow | moe down              |
    | ENTER      | delete label          |
    | a          | add label             |
    | e          | edit label            |
    | ESC        | back to switch boards |
