(function() {

  var AJAX_DONE = 4;
  var HTTP_OK = 200;
  var ENTER_KEYCODE = 13;

  var addButton;

  function initialize() {
    addButton = document.getElementById('add-button');
    addButton.addEventListener('click', addLabel);

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
    var labelField = document.getElementById('add-content');
    var label = labelField.value;
    var addURL = '/add?label=' + encodeURIComponent(label);
    addButton.disabled = true;
    getURL(addURL, function(err, id) {
      addButton.disabled = false;
      if (!err) {
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

  function registerRecordButton(button) {
    // TODO: register the click event, etc.
  }

  function registerDeleteButton(button) {
    var id = idForButton(button);
    button.addEventListener('click', function() {
      var url = '/delete?id=' + encodeURIComponent(id);
      getURL(url, function(err) {
        if (err) {
          showError(err);
        } else {
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

  function getURL(reqURL, callback) {
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
      if (xhr.readyState === AJAX_DONE) {
        if (xhr.status === HTTP_OK) {
          callback(null, xhr.responseText);
        } else {
          callback('GET failed with status: '+xhr.status, null);
        }
      }
    };
    xhr.open('GET', reqURL);
    xhr.send(null);
  }

  window.addEventListener('load', initialize);

})();
