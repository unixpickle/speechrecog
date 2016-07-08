// Code from https://github.com/unixpickle/jswav
(function() {

  function Recorder() {
    this.ondone = null;
    this.onerror = null;
    this.onstart = null;
    this.channels = 2;
    this._started = false;
    this._stopped = false;
    this._stream = null;
  }

  Recorder.prototype.start = function() {
    if (this._started) {
      throw new Error('Recorder was already started.');
    }
    this._started = true;
    getUserMedia(function(err, stream) {
      if (this._stopped) {
        return;
      }
      if (err !== null) {
        if (this.onerror !== null) {
          this.onerror(err);
        }
        return;
      }
      addStopMethod(stream);
      this._stream = stream;
      try {
        this._handleStream();
      } catch (e) {
        this._stream.stop();
        this._stopped = true;
        if (this.onerror !== null) {
          this.onerror(e);
        }
      }
    }.bind(this));
  };

  Recorder.prototype.stop = function() {
    if (!this._started) {
      throw new Error('Recorder was not started.');
    }
    if (this._stopped) {
      return;
    }
    this._stopped = true;
    if (this._stream !== null) {
      var stream = this._stream;
      this._stream.stop();
      // Firefox does not fire the onended event.
      setTimeout(function() {
        if (stream.onended) {
          stream.onended();
        }
      }, 500);
    }
  };

  Recorder.prototype._handleStream = function() {
    var context = getAudioContext();
    var source = context.createMediaStreamSource(this._stream);
    var wavNode = new window.jswav.WavNode(context, this.channels);
    source.connect(wavNode.node);
    wavNode.node.connect(context.destination);
    this._stream.onended = function() {
      this._stream.onended = null;
      source.disconnect(wavNode.node);
      wavNode.node.disconnect(context.destination);
      if (this.ondone !== null) {
        this.ondone(wavNode.sound());
      }
    }.bind(this);
    if (this.onstart !== null) {
      this.onstart();
    }
  };

  function getUserMedia(cb) {
    var gum = (navigator.getUserMedia || navigator.webkitGetUserMedia ||
      navigator.mozGetUserMedia || navigator.msGetUserMedia);
    if (!gum) {
      setTimeout(function() {
        cb('getUserMedia() is not available.', null);
      }, 10);
      return;
    }
    gum.call(navigator, {audio: true, video: false},
      function(stream) {
        cb(null, stream);
      },
      function(err) {
        cb(err, null);
      }
    );
  }

  function addStopMethod(stream) {
    if ('undefined' === typeof stream.stop) {
      stream.stop = function() {
        var tracks = this.getTracks();
        for (var i = 0, len = tracks.length; i < len; ++i) {
          tracks[i].stop();
        }
      };
    }
  }

  var reusableAudioContext = null;

  function getAudioContext() {
    if (reusableAudioContext !== null) {
      return reusableAudioContext;
    }
    var AudioContext = (window.AudioContext || window.webkitAudioContext);
    reusableAudioContext = new AudioContext();
    return reusableAudioContext;
  }

  if (!window.jswav) {
    window.jswav = {};
  }
  window.jswav.Recorder = Recorder;

})();
(function() {

  function Header(view) {
    this.view = view;
  }

  Header.prototype.getBitsPerSample = function() {
    return this.view.getUint16(34, true);
  };

  Header.prototype.getChannels = function() {
    return this.view.getUint16(22, true);
  };

  Header.prototype.getDuration = function() {
    return this.getSampleCount() / this.getSampleRate();
  };

  Header.prototype.getSampleCount = function() {
    var bps = this.getBitsPerSample() * this.getChannels() / 8;
    return this.view.getUint32(40, true) / bps;
  };

  Header.prototype.getSampleRate = function() {
    return this.view.getUint32(24, true);
  };

  Header.prototype.setDefaults = function() {
    this.view.setUint32(0, 0x46464952, true); // RIFF
    this.view.setUint32(8, 0x45564157, true); // WAVE
    this.view.setUint32(12, 0x20746d66, true); // "fmt "
    this.view.setUint32(16, 0x10, true); // length of "fmt"
    this.view.setUint16(20, 1, true); // format = PCM
    this.view.setUint32(36, 0x61746164, true); // "data"
  };

  Header.prototype.setFields = function(count, rate, bitsPerSample, channels) {
    totalSize = count * (bitsPerSample / 8) * channels;

    this.view.setUint32(4, totalSize + 36, true); // size of "RIFF"
    this.view.setUint16(22, channels, true); // channel count
    this.view.setUint32(24, rate, true); // sample rate
    this.view.setUint32(28, rate * channels * bitsPerSample / 8,
      true); // byte rate
    this.view.setUint16(32, bitsPerSample * channels / 8, true); // block align
    this.view.setUint16(34, bitsPerSample, true); // bits per sample
    this.view.setUint32(40, totalSize, true); // size of "data"
  };

  function Sound(buffer) {
    this.buffer = buffer;
    this._view = new DataView(buffer);
    this.header = new Header(this._view);
  }

  Sound.fromBase64 = function(str) {
    var raw = window.atob(str);
    var buffer = new ArrayBuffer(raw.length);
    var bytes = new Uint8Array(buffer);
    for (var i = 0; i < raw.length; ++i) {
      bytes[i] = raw.charCodeAt(i);
    }
    return new Sound(buffer);
  };

  Sound.prototype.average = function(start, end) {
    var startIdx = this.indexForTime(start);
    var endIdx = this.indexForTime(end);
    if (endIdx-startIdx === 0) {
      return 0;
    }
    var sum = 0;
    var channels = this.header.getChannels();
    for (var i = startIdx; i < endIdx; ++i) {
      for (var j = 0; j < channels; ++j) {
        sum += Math.abs(this.getSample(i, j));
      }
    }
    return sum / (channels*(endIdx-startIdx));
  };

  Sound.prototype.base64 = function() {
    var binary = '';
    var bytes = new Uint8Array(this.buffer);
    for (var i = 0, len = bytes.length; i < len; ++i) {
      binary += String.fromCharCode(bytes[i]);
    }
    return window.btoa(binary);
  };

  Sound.prototype.crop = function(start, end) {
    var startIdx = this.indexForTime(start);
    var endIdx = this.indexForTime(end);

    // Figure out a bunch of math
    var channels = this.header.getChannels();
    var bps = this.header.getBitsPerSample();
    var copyCount = endIdx - startIdx;
    var blockSize = channels * bps / 8;
    var copyBytes = blockSize * copyCount;

    // Create a new buffer
    var buffer = new ArrayBuffer(copyBytes + 44);
    var view = new DataView(buffer);

    // Setup the header
    var header = new Header(view);
    header.setDefaults();
    header.setFields(copyCount, this.header.getSampleRate(), bps, channels);

    // Copy the sample data
    var bufferSource = startIdx*blockSize + 44;
    for (var i = 0; i < copyBytes; ++i) {
      view.setUint8(i+44, this._view.getUint8(bufferSource+i));
    }

    return new Sound(buffer);
  };

  Sound.prototype.getSample = function(idx, channel) {
    if ('undefined' === typeof channel) {
      // Default value of channel is 0.
      channel = 0;
    }
    var bps = this.header.getBitsPerSample();
    var channels = this.header.getChannels();
    if (bps === 8) {
      var offset = 44 + idx*channels + channel;
      return (this._view.getUint8(offset)-0x80) / 0x80;
    } else if (bps === 16) {
      var offset = 44 + idx*channels*2 + channel*2;
      return this._view.getInt16(offset, true) / 0x8000;
    } else {
      return NaN;
    }
  };

  Sound.prototype.histogram = function(num) {
    var duration = this.header.getDuration();
    var timeSlice = duration / num;
    var result = [];
    for (var i = 0; i < num; ++i) {
      result.push(this.average(i*timeSlice, (i+1)*timeSlice));
    }
    return result;
  };

  Sound.prototype.indexForTime = function(time) {
    var samples = this.header.getSampleCount();
    var duration = this.header.getDuration();
    var rawIdx = Math.floor(samples * time / duration);
    return Math.min(Math.max(rawIdx, 0), samples);
  };

  if (!window.jswav) {
    window.jswav = {};
  }
  window.jswav.Sound = Sound;
  window.jswav.Header = Header;

})();
(function() {

  function WavNode(context, ch) {
    this.node = null;
    this._buffers = [];
    this._sampleCount = 0;
    this._sampleRate = 0;
    this._channels = 0;
    if (context.createScriptProcessor) {
      this.node = context.createScriptProcessor(1024, ch, ch);
    } else if (context.createJavaScriptNode) {
      this.node = context.createJavaScriptNode(1024, ch, ch);
    } else {
      throw new Error('No javascript processing node available.');
    }
    this.node.onaudioprocess = function(event) {
      var input = event.inputBuffer;
      if (this._sampleRate === 0) {
        this._sampleRate = Math.round(input.sampleRate);
      }
      if (this._channels === 0) {
        this._channels = input.numberOfChannels;
      }

      // Interleave the audio data
      var sampleCount = input.length;
      this._sampleCount += sampleCount;
      var buffer = new ArrayBuffer(sampleCount * this._channels * 2);
      var view = new DataView(buffer);
      var x = 0;
      for (var i = 0; i < sampleCount; ++i) {
        for (var j = 0; j < this._channels; ++j) {
          var value = Math.round(input.getChannelData(j)[i] * 0x8000);
          view.setInt16(x, value, true);
          x += 2;
        }
      }
      this._buffers.push(buffer);

      // If I don't do this, the entire thing backs up after a few buffers.
      event.outputBuffer = event.inputBuffer;
    }.bind(this);
  }

  WavNode.prototype.sound = function() {
    // Setup the buffer
    var buffer = new ArrayBuffer(44 + this._sampleCount*this._channels*2);
    var view = new DataView(buffer);

    // Setup the header
    var header = new window.jswav.Header(view);
    header.setDefaults();
    header.setFields(this._sampleCount, this._sampleRate, 16, this._channels);

    // Copy the raw data
    var byteIdx = 44;
    for (var i = 0; i < this._buffers.length; ++i) {
      var aBuffer = this._buffers[i];
      var aView = new DataView(aBuffer);
      var len = aBuffer.byteLength;
      for (var j = 0; j < len; ++j) {
        view.setUint8(byteIdx++, aView.getUint8(j));
      }
    }

    return new window.jswav.Sound(buffer);
  };

  if (!window.jswav) {
    window.jswav = {};
  }
  window.jswav.WavNode = WavNode;

})();
