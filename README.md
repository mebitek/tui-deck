# tui-deck
A TUI frontend for Nextcloud Deck app written in GO using the [Rich Interactive Widgets for Terminal UIs
](https://github.com/rivo/tview)

![image](https://github.com/mebitek/tui-deck/assets/1067967/4c1913be-09d0-4fea-bc67-19da89e2e9aa)

___

# features

* switch between boardsmebitek
* list cards
* edit card description
* move cards between stacksmebitek
* add/remove labels from cards
* theming

# planned features

* add/remove cards
* add/edit/delete stacks
* add/edit/delete boards
* manage comments
* manage attachments
* improve boot time with local storage data

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
    | ESC        | back to main view |


