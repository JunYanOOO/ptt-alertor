總人數：{{.Total}}
LINE: {{.Line}}, Messenger: {{.Messenger}}, Telegram: {{.Telegram}}, Discord: {{.Discord}}
BlockUser: {{.BlockUser}}, IdleUser: {{.IdleUser}}
User: {{.User}}, Room: {{.Room}}, Group: {{.Group}}
count(Board): {{.BoardCount}}{{if .Features.KeywordTracking}}, count(Keyword): {{.KeywordCount}}{{end}}{{if .Features.AuthorTracking}}, count(Author): {{.AuthorCount}}{{end}}{{if .Features.PushSumTracking}}, count(PushSum): {{.PushSumCount}}{{end}}
{{range .Users}}
{{.Profile.Account}}{{end}}
