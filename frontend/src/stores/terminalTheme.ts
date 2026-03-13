import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ITheme } from '@xterm/xterm'

export interface TerminalThemeDef {
  id: string
  name: string
  theme: ITheme
}

export const TERMINAL_THEMES: TerminalThemeDef[] = [
  {
    id: 'default',
    name: 'Default Dark',
    theme: {
      background: '#1e1e1e', foreground: '#d4d4d4', cursor: '#d4d4d4',
      selectionBackground: '#264f78',
      black: '#000000', red: '#cd3131', green: '#0dbc79', yellow: '#e5e510',
      blue: '#2472c8', magenta: '#bc3fbc', cyan: '#11a8cd', white: '#e5e5e5',
      brightBlack: '#666666', brightRed: '#f14c4c', brightGreen: '#23d18b',
      brightYellow: '#f5f543', brightBlue: '#3b8eea', brightMagenta: '#d670d6',
      brightCyan: '#29b8db', brightWhite: '#e5e5e5',
    },
  },
  {
    id: 'dracula',
    name: 'Dracula',
    theme: {
      background: '#282a36', foreground: '#f8f8f2', cursor: '#f8f8f2',
      selectionBackground: '#44475a',
      black: '#21222c', red: '#ff5555', green: '#50fa7b', yellow: '#f1fa8c',
      blue: '#bd93f9', magenta: '#ff79c6', cyan: '#8be9fd', white: '#f8f8f2',
      brightBlack: '#6272a4', brightRed: '#ff6e6e', brightGreen: '#69ff94',
      brightYellow: '#ffffa5', brightBlue: '#d6acff', brightMagenta: '#ff92df',
      brightCyan: '#a4ffff', brightWhite: '#ffffff',
    },
  },
  {
    id: 'monokai',
    name: 'Monokai',
    theme: {
      background: '#272822', foreground: '#f8f8f2', cursor: '#f8f8f0',
      selectionBackground: '#49483e',
      black: '#272822', red: '#f92672', green: '#a6e22e', yellow: '#f4bf75',
      blue: '#66d9ef', magenta: '#ae81ff', cyan: '#a1efe4', white: '#f8f8f2',
      brightBlack: '#75715e', brightRed: '#f92672', brightGreen: '#a6e22e',
      brightYellow: '#f4bf75', brightBlue: '#66d9ef', brightMagenta: '#ae81ff',
      brightCyan: '#a1efe4', brightWhite: '#f9f8f5',
    },
  },
  {
    id: 'solarized-dark',
    name: 'Solarized Dark',
    theme: {
      background: '#002b36', foreground: '#839496', cursor: '#839496',
      selectionBackground: '#073642',
      black: '#073642', red: '#dc322f', green: '#859900', yellow: '#b58900',
      blue: '#268bd2', magenta: '#d33682', cyan: '#2aa198', white: '#eee8d5',
      brightBlack: '#002b36', brightRed: '#cb4b16', brightGreen: '#586e75',
      brightYellow: '#657b83', brightBlue: '#839496', brightMagenta: '#6c71c4',
      brightCyan: '#93a1a1', brightWhite: '#fdf6e3',
    },
  },
  {
    id: 'solarized-light',
    name: 'Solarized Light',
    theme: {
      background: '#fdf6e3', foreground: '#657b83', cursor: '#586e75',
      selectionBackground: '#eee8d5',
      black: '#073642', red: '#dc322f', green: '#859900', yellow: '#b58900',
      blue: '#268bd2', magenta: '#d33682', cyan: '#2aa198', white: '#eee8d5',
      brightBlack: '#002b36', brightRed: '#cb4b16', brightGreen: '#586e75',
      brightYellow: '#657b83', brightBlue: '#839496', brightMagenta: '#6c71c4',
      brightCyan: '#93a1a1', brightWhite: '#fdf6e3',
    },
  },
  {
    id: 'nord',
    name: 'Nord',
    theme: {
      background: '#2e3440', foreground: '#d8dee9', cursor: '#d8dee9',
      selectionBackground: '#434c5e',
      black: '#3b4252', red: '#bf616a', green: '#a3be8c', yellow: '#ebcb8b',
      blue: '#81a1c1', magenta: '#b48ead', cyan: '#88c0d0', white: '#e5e9f0',
      brightBlack: '#4c566a', brightRed: '#bf616a', brightGreen: '#a3be8c',
      brightYellow: '#ebcb8b', brightBlue: '#81a1c1', brightMagenta: '#b48ead',
      brightCyan: '#8fbcbb', brightWhite: '#eceff4',
    },
  },
  {
    id: 'one-dark',
    name: 'One Dark',
    theme: {
      background: '#282c34', foreground: '#abb2bf', cursor: '#528bff',
      selectionBackground: '#3e4451',
      black: '#282c34', red: '#e06c75', green: '#98c379', yellow: '#e5c07b',
      blue: '#61afef', magenta: '#c678dd', cyan: '#56b6c2', white: '#abb2bf',
      brightBlack: '#5c6370', brightRed: '#e06c75', brightGreen: '#98c379',
      brightYellow: '#e5c07b', brightBlue: '#61afef', brightMagenta: '#c678dd',
      brightCyan: '#56b6c2', brightWhite: '#ffffff',
    },
  },
  {
    id: 'gruvbox',
    name: 'Gruvbox Dark',
    theme: {
      background: '#282828', foreground: '#ebdbb2', cursor: '#ebdbb2',
      selectionBackground: '#3c3836',
      black: '#282828', red: '#cc241d', green: '#98971a', yellow: '#d79921',
      blue: '#458588', magenta: '#b16286', cyan: '#689d6a', white: '#a89984',
      brightBlack: '#928374', brightRed: '#fb4934', brightGreen: '#b8bb26',
      brightYellow: '#fabd2f', brightBlue: '#83a598', brightMagenta: '#d3869b',
      brightCyan: '#8ec07c', brightWhite: '#ebdbb2',
    },
  },
  {
    id: 'material',
    name: 'Material Dark',
    theme: {
      background: '#212121', foreground: '#eeffff', cursor: '#ffcc02',
      selectionBackground: '#546e7a',
      black: '#212121', red: '#f07178', green: '#c3e88d', yellow: '#ffcb6b',
      blue: '#82aaff', magenta: '#c792ea', cyan: '#89ddff', white: '#eeffff',
      brightBlack: '#546e7a', brightRed: '#ff5370', brightGreen: '#c3e88d',
      brightYellow: '#ffcb6b', brightBlue: '#82aaff', brightMagenta: '#c792ea',
      brightCyan: '#89ddff', brightWhite: '#ffffff',
    },
  },
  {
    id: 'tomorrow-night',
    name: 'Tomorrow Night',
    theme: {
      background: '#1d1f21', foreground: '#c5c8c6', cursor: '#c5c8c6',
      selectionBackground: '#373b41',
      black: '#1d1f21', red: '#cc6666', green: '#b5bd68', yellow: '#f0c674',
      blue: '#81a2be', magenta: '#b294bb', cyan: '#8abeb7', white: '#c5c8c6',
      brightBlack: '#666666', brightRed: '#d54e53', brightGreen: '#b9ca4a',
      brightYellow: '#e7c547', brightBlue: '#7aa6da', brightMagenta: '#c397d8',
      brightCyan: '#70c0b1', brightWhite: '#eaeaea',
    },
  },
]

export const useTerminalThemeStore = defineStore('terminalTheme', () => {
  const STORAGE_KEY = 'terminal-theme'
  const currentId = ref(localStorage.getItem(STORAGE_KEY) ?? 'default')

  function current(): TerminalThemeDef {
    return TERMINAL_THEMES.find(t => t.id === currentId.value) ?? TERMINAL_THEMES[0]!
  }

  function apply(id: string) {
    currentId.value = id
    localStorage.setItem(STORAGE_KEY, id)
  }

  return { currentId, current, apply, themes: TERMINAL_THEMES }
})
