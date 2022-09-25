# Fantasm

A simple menu navigation VK bot with image support.

## Configuration

```yaml
# VK group token with messages access
vk_token: "vk1.a.ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-xyz"

# Templates for actions that can be used in sections for button actions
# Available colors for VK keyboard buttons: negative, positive, primary, secondary
action_templates:
  # An example of back action
  back:
    label: "Back"
    color: negative
    section: __back # magic name for returning to previous section
  to_start:
    label: "Main Menu"
    color: negative
    section: start # "start" is always the default section
  guide:
    label: "Guide"
    color: primary
    section: "{current}_guide" # {current} is being replaced with ID of current section
```

## Section description

You can create any number of YAML files describing sections in the ``sections`` directory, as well as in its subdirectories.

They will be parsed as they are all merged into one big YAML file.

Here's an example of section descriptions:

```yaml
# Available button colors: primary, secondary, negative

# start is always the default section
start:
  text: |
    Hello!
    We have buttons here!
  actions:
    - label: "Section 1"
      color: primary
      section: section1
    - label: "Section 2"
      color: secondary
      section: section2

section1:
  text: |
    Section 1!
  actions:
    - label: "Subsection 1"
      color: primary
      section: subsection1
    - template: guide
      color: secondary # We can replace keys from template
    - template: back

section1_guide:
  text: |
    Section 1 Guide
  actions:
    - template: back

subsection1:
  text: |
    Subsection 1!
  actions:
    - label: "Section 2"
      color: primary
      section: section2
    - template: to_start
    - template: back

section2:
  text: |
    Section 2!
  actions:
    - template: to_start
    - template: back
```