$word = New-Object -ComObject Word.Application
$word.Visible = $false
$doc = $word.Documents.Open("c:\Users\dima\Desktop\Смирнов матьего\Требования_АС_Чистое_Небо_ласт_апдейт_на_паре.docx")
$text = $doc.Content.Text
$doc.Close()
$word.Quit()
[System.Runtime.Interopservices.Marshal]::ReleaseComObject($word) | Out-Null
$text | Out-File -FilePath "c:\Users\dima\Desktop\Смирнов матьего\requirements_text.txt" -Encoding UTF8
