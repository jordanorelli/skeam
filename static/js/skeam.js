// -----------------------------------------------------------------------------
// input handling
// -----------------------------------------------------------------------------

var isSendMessage = function(event) {
  return event.keyCode == 13 && event.shiftKey;
}

var InputHandler = function(selector) {
  this.elem = $(selector);
  this.elem.keydown(_.bind(this.keydown, this));
};

InputHandler.prototype.keydown = function(event) {
  if (isSendMessage(event)) {
    return this.handleSend(event);
  }
};

InputHandler.prototype.handleSend = function(event) {
  event.preventDefault();
  event.stopPropagation();
  this.sendMessage(event.target.value);
  this.clear();
};

InputHandler.prototype.sendMessage = function(message) {
  var event = document.createEvent("Event");
  event.initEvent("sendMsg", true, true);
  event.message = message;
  document.dispatchEvent(event);
};

InputHandler.prototype.clear = function() {
  this.elem.val('');
}

// -----------------------------------------------------------------------------
// connection handling
// -----------------------------------------------------------------------------

var buildWsPath = function(wsPath) {
  if (window.location.port) {
    return "ws://" + window.location.hostname + ":" + window.location.port + wsPath;
  }
  return "ws://" + window.location.hostname + wsPath;
}

var ConnectionHandler = function(wsPath) {
  this.c = new WebSocket(buildWsPath(wsPath));
  this.c.onopen = _.bind(this.onopen, this);
  this.c.onclose = _.bind(this.onclose, this);
  this.c.onerror = _.bind(this.onerror, this);
  this.c.onmessage = _.bind(this.onmessage, this);
};

ConnectionHandler.prototype.onopen = function(event) {
  console.log({open: event});
};

ConnectionHandler.prototype.onclose = function(event) {
  console.log({close: event});
};

ConnectionHandler.prototype.onmessage = function(event) {
  this.broadcastMessageReceived(event.data);
};

ConnectionHandler.prototype.broadcastMessageReceived = function(message) {
  var event = document.createEvent("Event");
  event.initEvent("receiveResponse", true, true);
  event.message = message;
  document.dispatchEvent(event);
}

ConnectionHandler.prototype.onerror = function(event) {
  console.log({error: event});
};

ConnectionHandler.prototype.sendMsg = function(message) {
  this.c.send(message);
};

// -----------------------------------------------------------------------------
// response rendering
// -----------------------------------------------------------------------------

var MessageDisplay = function(selector, templateSelector, errorTemplateSelector) {
  this.elem = $(selector);
  this.renderMessage = _.template($(templateSelector).html());
  this.renderError = _.template($(errorTemplateSelector).html());
}

MessageDisplay.prototype.addMessage = function(rawmessage) {
  var message = JSON.parse(rawmessage);
  if (message.is_error) {
    var rendered = this.renderError(message);
  } else {
    var rendered = this.renderMessage(message);
  }
  this.elem.append(rendered);
};

var Skeam = function(config) {
  this.inputHandler = new InputHandler(config.inputSelector);
  this.messageDisplay = new MessageDisplay(config.outputSelector,
                                           config.messageTemplateSelector,
                                           config.errorTemplateSelector);
  this.conn = new ConnectionHandler(config.wsPath);
  document.addEventListener("sendMsg", _.bind(this.sendMsg, this), false);
  document.addEventListener("receiveResponse", _.bind(this.receiveResponse, this), false);
};

Skeam.prototype.sendMsg = function(event) {
  this.conn.sendMsg(event.message + "\n");
}

Skeam.prototype.receiveResponse = function(event) {
  this.messageDisplay.addMessage(event.message);
}
