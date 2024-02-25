# tui-deck
A TUI frontend for Nextcloud [Deck](https://github.com/nextcloud/deck) app written in GO using the [Rich Interactive Widgets for Terminal UIs
](https://github.com/rivo/tview)

### [AUR package](https://aur.archlinux.org/packages/tui-deck)

![image](https://github.com/mebitek/tui-deck/assets/1067967/4c1913be-09d0-4fea-bc67-19da89e2e9aa)

___

# features

* switch between boards
* list cards
* edit card description, title, due date
* move cards between stacks
* add/remove labels from cards
* add/edit/remove stacks
* add/edit/remove boards
* add/edit/remove boards labels
* basic markdown viewer
* assign users to card
* comments
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
* links

# planned features

- [ ] manage attachments

# configuration

on first start, the application will create a default config.json file in $HOME/.config/tui-deck directory

```
{
  "username": "",
  "password": "",
  "url": "https://nextcloud.example.com",
  "color": "#BF40BF"
  "color": "#BF40BF",
  "insecure": false # Set to true if you're using self-signed certificates or you need to bypass certificate verification
  "configDir": "$HOME/.config/tui-deck/"
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
    | l        | edit card labels      |
    | u        | edit card users       |
    | t        | edit card title       |
    | c        | view comments         |
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
    | TAB        | switch between card labels and available board labels list                                       |
    | ENTER      | if card label has been selected, delete it. if available label has been selected, add it to card |
    | ESC        | back to view card                                                                                |

* edit card users

    | function   | key                                                                                            |
    |------------|------------------------------------------------------------------------------------------------|
    | up arrow   | move up                                                                                        |
    | down arrow | move down                                                                                      |
    | TAB        | switch between card users and available board users list                                       |
    | ENTER      | if card user has been selected, delete it. if available user has been selected, add it to card |
    | ESC        | back to view card                                                                              |

* view comments 

    | function   | key                       |
    |------------|---------------------------|
    | up arrow   | move up                   |
    | down arrow | move down                 |
    | a          | add comment               |
    | r          | reply to selected comment |
    | e          | edit comment              |
    | d          | delete selected comment   | 
    | ESC        | back to view card         |

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
