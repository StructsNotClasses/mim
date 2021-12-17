module github.com/StructsNotClasses/musicplayer

go 1.17

replace musicplayer/github.com/d5/tengo/v2 v2.0.0 => ./tengo

//require github.com/rthornton128/goncurses v0.0.0-20211122162138-db8d4cdb33a9
//require musicplayer/tengo v0.0.0

replace github.com/StructsNotClasses/tengotango => /home/pugpugpugs/tengotango/
replace github.com/StructsNotclasses/musicplayer => /mnt/music/music_player/

require (
	github.com/StructsNotClasses/tengotango v1.24.8 // indirect
	github.com/rthornton128/goncurses v0.0.0-20211122162138-db8d4cdb33a9 // indirect
)
