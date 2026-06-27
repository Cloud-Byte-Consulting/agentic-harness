# Image Generation Gateway

Generates or edits images through a single API (OpenRouter is a good choice) with one command and zero per-call setup. The skill stores your saved preferences — default model, output directory, default size — so "generate an image of X" just works. It captures the current request shape of the API: which fields the endpoint expects, which model IDs are live, what each model costs per image, and the gotchas you've already hit. Other skills reference this one instead of writing their own API code.

## Why Build It
Image APIs change constantly, and every agent session that improvises an API call repeats old mistakes. Centralizing image generation in one skill means you fix an API change once, and every workflow that generates images inherits the fix. This is the clearest example of a skill as a shared primitive: at least three other skills in this library call it rather than reimplementing it.

## What You Need


## Prompt / Setup
```xml
<prompt>
  <task>
    Create a new AI coding-agent skill called "image-gateway".
  </task>

  <storage>
    Store the skill wherever this harness loads skills from, such as
    ~/.claude/skills/image-gateway/SKILL.md or ~/.codex/skills/image-gateway/SKILL.md.
  </storage>

  <job>
    The skill generates or edits images through the OpenRouter API with one command.
    It should use saved preferences so routine image requests do not require per-call setup.
  </job>

  <inputs_to_collect>
    <input>Preferred default image model.</input>
    <input>Default output directory.</input>
    <input>Where the OpenRouter API key lives. It must be read from an env file, never written into the skill.</input>
  </inputs_to_collect>

  <requirements>
    <requirement>Define trigger conditions for direct image-generation requests and for other skills that need image generation.</requirement>
    <requirement>Document the current OpenRouter image API request shape.</requirement>
    <requirement>Include a working curl or script example that reads the key from the env file.</requirement>
    <requirement>Store the collected preferences in the skill.</requirement>
    <requirement>Include per-image cost notes for the selected default model.</requirement>
    <requirement>Tell other skills to call this skill instead of writing their own image API code.</requirement>
  </requirements>

  <verification>
    Generate one image with the saved defaults, show me the result, then update the skill with anything learned from the test.
  </verification>
</prompt>
```
