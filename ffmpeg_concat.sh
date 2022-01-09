# https://newbedev.com/concatenate-multiple-wav-files-using-single-command-without-extra-file

ffmpeg -y -i a.wav -i i.wav -i j.wav \
-filter_complex '[0:0][1:0][2:0]concat=n=3:v=0:a=1[out]' \
-map '[out]' jjj.wav
