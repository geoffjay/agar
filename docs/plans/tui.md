# TUI Plan

These are all concepts that need to be implemented.

## Question

A panel component that shows the current question and allows the user to answer it in the panel that's provided. It
should support all question/input types.

## Prompt

A panel component that allows the user to input a prompt for use by the application.

### Prompt History Recall

This component should allow the user to cycle through previous inputs that have been made, this will require a
connection point to the state of the application where all inputs that are submitted are stored so that they can be
recalled from the history. This component requires a plan beyone just the implementation of the component itself. This
requires session tracking in the application so the ability to traverse history should probably be optional.

### Slash Commands

The prompt component should support slash commands.

## Dialog

The dialog component should be able to be displayed on top of all other components.
