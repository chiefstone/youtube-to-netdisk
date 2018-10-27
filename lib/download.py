from youtube_dl import YoutubeDL
import sys
import logging
import os

if len(sys.argv) <= 1:
    logging.error("No url specified")
    exit()

url = sys.argv[1]

print(url)

with YoutubeDL({
        'format': "bestvideo+bestaudio/best",
        'outtmpl': '%(title)s.%(ext)s',
        'writesubtitles': True,
        'allsubtitles': True,
    }) as ydl:
    info_dict = ydl.extract_info(url, download=True)
    fn = ydl.prepare_filename(info_dict)

if os.path.exists(fn):
    print("fn:\"" + fn + "\"")
elif os.path.exists(os.path.splitext(fn)[0] + '.mkv'):
    print("fn:\"" + os.path.splitext(fn)[0] + '.mkv' + "\"")
else:
    print("Error: file doesn't exist")