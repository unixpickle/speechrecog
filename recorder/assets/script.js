(function() {

  var AJAX_DONE = 4;
  var HTTP_OK = 200;
  var ENTER_KEYCODE = 13;
  var ESCAPE_KEYCODE = 27;

  var addButton = null;
  var currentRecorder = null;
  var currentRecordButton = null;

  function initialize() {
    addButton = document.getElementById('add-button');
    addButton.addEventListener('click', addLabel);

    window.addEventListener('keyup', function(e) {
      if (e.which === ESCAPE_KEYCODE) {
        e.preventDefault();
        cancelRecording();
      }
    });

    document.getElementById('add-content').addEventListener('keyup', function(e) {
      e.preventDefault();
      if (e.which === ENTER_KEYCODE) {
        addLabel();
      }
    });

    var buttonClasses = ['record-button', 'delete-button'];
    var buttonRegistrars = [registerRecordButton, registerDeleteButton];
    for (var i = 0, len = buttonClasses.length; i < len; ++i) {
      var className = buttonClasses[i];
      var reg = buttonRegistrars[i];
      var buttons = document.getElementsByClassName(className);
      for (var j = 0, len1 = buttons.length; j < len1; ++j) {
        reg(buttons[j]);
      }
    }
  }

  function addLabel() {
    cancelRecording();
    var labelField = document.getElementById('add-content');
    var label = labelField.value;
    var addURL = '/add?label=' + encodeURIComponent(label);
    addButton.disabled = true;
    getURL(addURL, function(err, id) {
      addButton.disabled = false;
      if (!err) {
        cancelRecording();
        labelField.value = null;
        addNewRow(id, label);
      } else {
        showError(err);
      }
    });
  }

  function showError(err) {
    alert(err);
  }

  function addNewRow(id, label) {
    var element = document.createElement('tr');
    element.setAttribute('label-id', id);

    var labelCol = document.createElement('td');
    labelCol.textContent = label;

    var recordCol = document.createElement('td');
    var recordButton = document.createElement('button');
    recordButton.className = 'record-button';
    recordButton.textContent = 'Record';
    recordCol.appendChild(recordButton);

    var deleteCol = document.createElement('td');
    var deleteButton = document.createElement('button');
    deleteButton.className = 'delete-button';
    deleteButton.textContent = 'Delete';
    deleteCol.appendChild(deleteButton);

    element.appendChild(labelCol);
    element.appendChild(recordCol);
    element.appendChild(deleteCol);

    document.getElementById('samples-body').appendChild(element);
    registerRecordButton(recordButton);
    registerDeleteButton(deleteButton);
  }

  function showAudioInRow(row) {
    var id = row.getAttribute('label-id');
    var oldCol = row.getElementsByTagName('td')[1];

    var newCol = document.createElement('td');
    var audioTag = document.createElement('audio');
    audioTag.controls = true;
    audioTag.preload = 'none';
    var sourceTag = document.createElement('source');
    sourceTag.src = '/recording.wav?id=' + encodeURIComponent(id);
    sourceTag.type = 'audio/x-wav';
    audioTag.appendChild(sourceTag);
    newCol.appendChild(audioTag);

    row.insertBefore(newCol, oldCol);
    row.removeChild(oldCol);
  }

  function registerRecordButton(button) {
    var id = idForButton(button);
    button.addEventListener('click', function() {
      if (button.textContent === 'Done') {
        currentRecorder.stop();
        button.textContent = 'Record';
        return;
      }
      cancelRecording();
      button.textContent = 'Done';
      currentRecordButton = button;
      currentRecorder = new window.jswav.Recorder();
      currentRecorder.ondone = function(sound) {
        currentRecorder = null;
        currentRecordButton = null;
        button.textContent = 'Record';
        uploadRecording(id, sound, function(err, data) {
          if (err) {
            showError(err);
          } else {
            showAudioInRow(rowForButton(button));
          }
        });
      };
      currentRecorder.onerror = function(err) {
        button.textContent = 'Record';
        currentRecorder = null;
        currentRecordButton = null;
        showError(err);
      };
      currentRecorder.start();
    });
  }

  function registerDeleteButton(button) {
    var id = idForButton(button);
    button.addEventListener('click', function() {
      cancelRecording();
      var url = '/delete?id=' + encodeURIComponent(id);
      getURL(url, function(err) {
        if (err) {
          showError(err);
        } else {
          cancelRecording();
          var row = rowForButton(button);
          row.parentElement.removeChild(row);
        }
      });
    });
  }

  function rowForButton(button) {
    return button.parentElement.parentElement;
  }

  function idForButton(button) {
    return rowForButton(button).getAttribute('label-id');
  }

  function cancelRecording() {
    if (currentRecorder !== null) {
      currentRecorder.ondone = null;
      currentRecorder.onerror = null;
      currentRecorder.stop();
      currentRecorder = null;
      currentRecordButton.textContent = 'Record';
      currentRecordButton = null;
    }
  }

  function uploadRecording(id, sound, callback) {
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
      if (xhr.readyState === AJAX_DONE) {
        if (xhr.status === HTTP_OK) {
          callback(null, xhr.responseText);
        } else {
          callback('Error '+xhr.status+': '+xhr.responseText, null);
        }
      }
    };
    xhr.open('POST', '/upload?id='+encodeURIComponent(id));
    xhr.send(sound.base64());
  }

  function getURL(reqURL, callback) {
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
      if (xhr.readyState === AJAX_DONE) {
        if (xhr.status === HTTP_OK) {
          callback(null, xhr.responseText);
        } else {
          callback('Error '+xhr.status+': '+xhr.responseText, null);
        }
      }
    };
    xhr.open('GET', reqURL);
    xhr.send(null);
  }

  window.addEventListener('load', initialize);

})();
