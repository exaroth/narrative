<h2 align="center">
<img align="center" width="70" height="70" src="https://github.com/user-attachments/assets/d4250f20-f0d2-4832-9415-b72e99aa3af7" alt="Narrative"><br/>
<br/>
Narrative
</h2>

Narrative streams and converts English language e-books and other text sources for your listening pleasure.

- Uses lightweight [KittenTTS](https://github.com/KittenML/KittenTTS) models for Text-to-Speech conversion, no need for GPU or AI account to run.
- Supports multiple input formats including __EPUB__, __MOBI__, __AZW3__ , __HTML__, __Markdown__ and plain text.
- Chapter Selection, Bookmarking support and Playback controls included.
- Built-in Text to MP3 converter.
- Voice selection.
- Built in phoneme editor/debugger.
- Proudly crafted without use of AI.

## Installation

TODO:

## Usage

```
Usage: narrative [OPTIONS...] TEXT_SOURCE
OPTIONS:
	-c --convert <source_path>  Convert text source to mp3 file.
	-v --voice <voice_name>     Set voice for playback.
	--list-voices               List available voices.
	-m --select-model <name>    Run Narrative using given model.
	--add-model <name>          Download and select KittenTTS model.
	--list-models               List available KittenTTS model information.
	-l --select-lib <lib>       Run narrative with provided library.
	--add-lib <lib>             Download and select ONNX lib with given name.
	--list-libs                 List available ONNX library for OS/Arch.
	-s --speed                  Select playback speed.
	--ext-dict                  Path to optional dictionary to override default phonemes.
	--init                      Reinitialize Narrative configuration and models.
	--help                      Print help.
```


Executing `narrative` for the first time will prompt for model selection, program will autoconfigure itself afterwards.

To add new text source simply pass path or url pointing to it as first argument.

### Switching models/libraries

By default Narrative will use CPU based ONNX library, if you want to use your PC's GPU for inference you can switch library by adding it via `--add-lib` argument, to view available libraries use `--list-libs` command. Library selection is currently only supported on `linux/amd64` architectures.

Use `--add-model` to add/switch TTS model in use.

## Packages

Narrative incorporates set of packages that can be used independently from the main application:

- **Kitten** - [https://github.com/exaroth/narrative/tree/main/pkg/kitten](https://github.com/exaroth/narrative/tree/main/pkg/kitten) - Used for direct interaction with KittenTTS models such as tokenization and inference.
- **Preprocessor** - [https://github.com/exaroth/narrative/tree/main/pkg/preprocessor](https://github.com/exaroth/narrative/tree/main/pkg/preprocessor) - Prepares input strings for phonemization - eg. does time/currency unit expansion and conversion of numerical values into human readable formats, unicode normalization and a lot more.
- **Phonemizer** - [https://github.com/exaroth/narrative/tree/main/pkg/phonemizer](https://github.com/exaroth/narrative/tree/main/pkg/phonemizer) - Module for phonemizing inputs, it's based on the [Goruut](https://github.com/neurlang/goruut) library, involving dictionary lookups, phoneme inferrence, punctuation splitting etc.
- **Reader** - [https://github.com/exaroth/narrative/tree/main/pkg/reader](https://github.com/exaroth/narrative/tree/main/pkg/reader) - Module used for reading and conversion of various text formats for usage within Narrative.


## Debugger

// todo
Narrative comes with debugger helpful for editing, modifying and adding phonemes used by Narrative. You can access it with F2 key from within main application or build standalone instance by executing `make build-debugger`. Running debugger will create new dir `narrative-debugger` in the current working directory, inside it  will create 2 files:

- `aux_dict.csv` - Contains phoneme overrides, these will be used instead of phonemes found in default dictionary or inferred by Narrative. Phonemes edited/added by debugger will be automatically added here.
- `missing_dict.csv` - Contains words that were missing in the built-in dictionary and were inferred by the Narrative.

In order to add edited phonemes to default dictionary you can use `merge-dicts` utility (eg. by running `just merge-dicts` assuming you are running debugger from within Narrative repo).

## Acknowledgments

Narrative would not be possible without these projects:

- **KitenTTS** [https://github.com/KittenML/KittenTTS](https://github.com/KittenML/KittenTTS)
- **Goruut** [https://github.com/neurlang/goruut](https://github.com/neurlang/goruut)
- **KindleUnpack** [https://github.com/kevinhendricks/KindleUnpack](https://github.com/kevinhendricks/KindleUnpack)
- **Bubbletea** [https://github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
- **Beep** [https://github.com/faiface/beep](https://github.com/faiface/beep)
