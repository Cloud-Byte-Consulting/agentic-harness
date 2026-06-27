# Media Transcription

Transcribes local audio or video files with a transcription API (AssemblyAI is a strong default) and packages the output into reusable artifacts: a clean readable Markdown transcript, word-level timestamps, semantic chapters, and speaker labels. The skill captures the current API request shape — including newer fields the docs bury — so transcription work never repeats old API mistakes. The output artifacts are deliberately designed to feed other skills: editing workflows, research synthesis, and content generation all start from these files.

## Why Build It
Transcripts are the universal input format for media work. Once a video or recording is a timestamped transcript, your agent can edit it, summarize it, fact-check it, extract clips from it, and generate graphics for it. Almost every media runbook in this library starts here. Getting the packaging right once — consistent filenames, chapters, timestamps — is what makes the downstream skills composable.

## What You Need


## Prompt / Setup
```xml
<prompt>
  <task>
    Create a new skill for my AI coding agent called "media-transcription", stored
wherever my harness loads skills from.

The skill's job: take a local audio or video file path and produce a complete
transcription package using the AssemblyAI API.

Before writing it, interview me for: where my AssemblyAI API key should live (env
file), and where transcription outputs should be saved (suggest a folder convention
next to the source media).

The skill must include: (1) trigger conditions — any time I give you a media file and
ask for a transcript, captions, chapters, or "make this searchable"; (2) the current
AssemblyAI request shape including the speech model field, with a working script that
reads the key from the env file; (3) a standard output package: readable Markdown
transcript, word-level timestamp JSON, semantic chapters, and speaker labels, all with
consistent filenames; (4) an ffmpeg step to extract audio from video first when
needed; (5) a note that these artifacts are inputs for editing and research skills, so
the format must stay consistent.

After writing it, test it on a short audio file and show me the output package.
  </task>
</prompt>
```
