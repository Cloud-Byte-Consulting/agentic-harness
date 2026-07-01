# vhs tape — financial tri-state policy demo. Render: vhs demo.tape
# The agent proposes tool calls; the policy returns the gateway verdict (200 / 428 / 403)
# and writes an audit record. The tail shows what one audit record looks like.
Output demo.gif

Set Shell "bash"
Set FontSize 18
Set Width 1100
Set Height 720
Set Padding 18
Set Theme "Catppuccin Mocha"
Set TypingSpeed 55ms

Hide
Type "rm -f decisions.jsonl"
Enter
Show

Type "# the LLM proposes a tool call — the policy decides, and every decision is audited."
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
Sleep 2.5s

Type "# every decision is appended to an audit log — what a risk reviewer reads:"
Enter
Sleep 800ms
Type "tail -1 decisions.jsonl | python3 -m json.tool"
Sleep 400ms
Enter
Sleep 4s
