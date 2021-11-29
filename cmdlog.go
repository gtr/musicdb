package main

import "fmt"

type Command struct {
	Method    string
	Arguments []string
}

type LogEntry struct {
	Command *Command
	Term    int
}

type CommandLog struct {
	Entries []LogEntry
}

func (l *CommandLog) AppendEntry(entry *LogEntry) {
	l.Entries = append(l.Entries, *entry)
}

func Reconstruct(db *AlbumDB, log *CommandLog) {
	for _, entry := range log.Entries {
		cmd := entry.Command
		if cmd.Method == "AddAlbum" {
			if len(cmd.Arguments) == 4 {
				db.AddAlbum(cmd.Arguments[0], cmd.Arguments[1], cmd.Arguments[2], cmd.Arguments[3])
			} else {
				fmt.Println("Invalid arguments for AddAlbum")
			}
		} else if cmd.Method == "EditAlbum" {
			if len(cmd.Arguments) == 5 {
				db.EditAlbum(cmd.Arguments[0], cmd.Arguments[1], cmd.Arguments[2], cmd.Arguments[3], cmd.Arguments[4])
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
}
