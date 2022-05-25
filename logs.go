package main

import "fmt"

// ================================ COMMAND LOG ===============================

// Command represents a command to be executed to our in-memory database.
type Command struct {
	Method    string
	Arguments []string
}

// LogEntry represents an entry in our log, consisting of a command and a term.
type LogEntry struct {
	Command *Command
	Term    int
}

// CommandLog represents a log of commands in our consensus module. When
// applied sequentially to our in-memory database, it should result in a
// reproducable state.
type CommandLog struct {
	Entries []LogEntry
}

// AppendEntry appends a new LogEntry into our log.
func (l *CommandLog) AppendEntry(entry *LogEntry) {
	l.Entries = append(l.Entries, *entry)
}

// applyCommand applied a given command to our in-memory database.
func applyCommand(db *AlbumDB, entry *LogEntry) {
	cmd := entry.Command
	if cmd.Method == "AddAlbum" {
		if len(cmd.Arguments) == 4 {
			db.AddAlbum(cmd.Arguments[0],
				cmd.Arguments[1],
				cmd.Arguments[2],
				cmd.Arguments[3])
		} else {
			fmt.Println("Invalid arguments for AddAlbum")
		}
	} else if cmd.Method == "EditAlbum" {
		if len(cmd.Arguments) == 5 {
			db.EditAlbum(cmd.Arguments[0],
				cmd.Arguments[1],
				cmd.Arguments[2],
				cmd.Arguments[3],
				cmd.Arguments[4])
		} else {
			fmt.Println("Invalid arguments for AddAlbum")
		}
	} else if cmd.Method == "RemoveAlbum" {
		if len(cmd.Arguments) == 1 {
			db.RemoveAlbum(cmd.Arguments[0])
		} else {
			fmt.Println("Invalid arguments for RemoveAlbum")
		}
	} else {
		fmt.Println("Unknown command")
	}
}

func Reconstruct(db *AlbumDB, log *CommandLog) {
	for _, entry := range log.Entries {
		applyCommand(db, &entry)
	}
}

// ================================ COMMIT LOG ================================

// EntryToCommit represents an entry (simillar to LogEntry) for which a
// consensus has been reached by a quorum and is ready to be committed.
type EntryToCommit struct {
	Command *Command
	Term    int
	Index   int
}
