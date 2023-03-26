# ChatGPT TUI

[![asciicast](https://asciinema.org/a/6iE1fbCZR5YeZlle9CujTbpTB.svg)](https://asciinema.org/a/6iE1fbCZR5YeZlle9CujTbpTB)

> Just found [j178/chatgpt](https://github.com/j178/chatgpt), which is a better
> implementation of this project. I won't maintain this project anymore, haha

## Install

```bash
go install github.com/Yukaii/chatgpt-tui@latest
```

## Usage

```bash
chatgpt-tui
```

## Screenshots

![screenshot](./docs/assets/screenshot.png)

## Roadmap

- [x] Basic communiation with OpenAI API
- [x] Basic Layout
- [x] Stream API response (preventing request timeout)
- [ ] Save conversation to file
- [ ] CLI option with [cobra](https://github.com/spf13/cobra)
- [ ] configuration with [viper](https://github.com/spf13/viper)
- [ ] bubblezone mouse support
