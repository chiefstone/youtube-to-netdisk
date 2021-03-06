from youtube_dl import YoutubeDL
import sys
import os
import json

if len(sys.argv) <= 1:
    exit()

url = sys.argv[1]

with YoutubeDL({
        'format': "bestvideo+bestaudio/best",
        'outtmpl': '%(title)s.%(ext)s',
        'writesubtitles': True,
        'allsubtitles': True,
        'quiet': True,
        'no_warnings': True,
    }) as ydl:
    info_dict = ydl.extract_info(url, download=False)
    print(json.dumps(info_dict))
