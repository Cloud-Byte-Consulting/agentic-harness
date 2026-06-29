# vhs tape — healthcare tri-state policy demo. Render: vhs demo.tape
# The agent proposes tool calls; the policy returns the gateway verdict (200 / 428 / 403)
# and writes an audit record. The tail shows what one audit record looks like.
Output demo.gif

Set Shell "bash"
Set FontSize 18
Set Width 1240
Set Height 900
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

Type "python3 selfcheck.py read_care_guidelines"
Sleep 400ms
Enter
Sleep 2s

Type "python3 selfcheck.py get_patient_record"
Sleep 400ms
Enter
Sleep 2s

Type "python3 selfcheck.py order_medication"
Sleep 400ms
Enter
Sleep 2s

Type "# fail-closed: an unknown tool is denied by default"
Enter
Sleep 600ms
Type "python3 selfcheck.py exfiltrate_patient_charts"
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
