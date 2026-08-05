/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package style

// Escape sequences start with ESC (ASCII 27 / 0x1B / 033). Most are longer than two chars and start with ESC+[.
// That prefix is the Control Sequence Introducer (CSI), usually written \033[ or \e[.

/*
\033[0m reset all attributes
      \033[1m bold / high intensity
      \033[4m underline
      \033[5m blink
      \033[7m reverse video
      \033[8m conceal
      \033[30m to \33[37m set foreground color
      \033[40m to \33[47m set background color
      \033[nA move cursor up n lines
      \033[nB move cursor down n lines
      \033[nC move cursor right n columns
      \033[nD move cursor left n columns
      \033[y;xH set cursor position
      \033[2J clear screen
      \033[K clear from cursor to end of line
      \033[s save cursor position
      \033[u restore cursor position
      \033[?25l hide cursor
      \033[?25h show cursor
*/
/*
Background color range: 40----49
      40: black
      41: dark red
      42: green
      43: yellow
      44: blue
      45: purple
      46: dark green
      47: white

      Foreground color: 30-----------39
      30: black
      31: red
      32: green
      33: yellow
      34: blue
      35: purple
      36: dark green
      37: white
*/
