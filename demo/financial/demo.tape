# vhs tape — financial tri-state policy demo. Render: vhs demo.tape
# Produces demo.gif: the offline self-check proving allow / require_approval / deny.
Output demo.gif

Set Shell "bash"
Set FontSize 18
Set Width 1100
Set Height 620
Set Padding 18
Set Theme "Catppuccin Mocha"
Set TypingSpeed 45ms

Type "# financial tri-state policy — the LLM proposes, OPA decides"
Enter
Sleep 1s

Type "python3 selfcheck.py"
Sleep 500ms
Enter
Sleep 3.5s

Type "# allow -> forward · require_approval -> 428 ASK · deny -> 403"
Enter
Sleep 2.5s
