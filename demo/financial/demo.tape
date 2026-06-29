# vhs tape — financial tri-state policy demo. Render: vhs demo.tape
# The agent proposes tool calls; the policy returns the gateway verdict (200 / 428 / 403).
Output demo.gif

Set Shell "bash"
Set FontSize 18
Set Width 1100
Set Height 720
Set Padding 18
Set Theme "Catppuccin Mocha"
Set TypingSpeed 55ms

Type "# the LLM proposes a tool call — the policy decides. each call returns a gateway verdict."
Enter
Sleep 1.5s

Type "python3 selfcheck.py read_education_content"
Sleep 400ms
Enter
Sleep 2.5s

Type "python3 selfcheck.py get_member_account"
Sleep 400ms
Enter
Sleep 2.5s

Type "python3 selfcheck.py send_external_message"
Sleep 400ms
Enter
Sleep 2.5s

Type "python3 selfcheck.py transfer_funds"
Sleep 400ms
Enter
Sleep 2.5s

Type "# fail-closed: an unknown tool gets denied by default"
Enter
Sleep 800ms
Type "python3 selfcheck.py exfiltrate_all_accounts"
Sleep 400ms
Enter
Sleep 3s
