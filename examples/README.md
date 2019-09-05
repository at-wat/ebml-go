# ebml-go examples

## rtp-to-webm

Receive RTP VP8 stream UDP packets and pack it to WebM file.

1. Run the following command.
    ```shell
    $ cd rtp-to-webm
    $ go build .
    $ ./rtp-to-webm
    ```
2. Send RTP stream to `./rtp-to-webm` using GStreamer.
    ```shell
    $ gst-launch-1.0 videotestsrc \
        ! video/x-raw,width=320,height=240,framerate=30/1 \
        ! vp8enc target-bitrate=4000 \
        ! rtpvp8pay ! udpsink host=localhost port=4000
    ```
3. Check out `test.webm` generated at the current directory.


## webm-roundtrip

Read WebM file, parse, and write back to the file.

1. Run the following command to read `sample.webm`.
    ```shell
    $ cd webm-roundtrip
    $ go build .
    $ ./webm-roundtrip
    ```
2. Contents (EBML document) of the file is shown to the stdout.
3. Check out `copy.webm` generated at the current directory. It should be playable and seekable as same as `sample.webm`.
