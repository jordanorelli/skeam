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
  this.broadcastClose();
};

ConnectionHandler.prototype.broadcastClose = function() {
  var event = document.createEvent("Event");
  event.initEvent("wsClosed", true, true);
  document.dispatchEvent(event);
}

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

var MessageDisplay = function(selector, requestTemplateSelector, responseTemplateSelector, errorTemplateSelector, sysTemplateSelector) {
  this.elem = $(selector);
  this.renderRequest = _.template($(requestTemplateSelector).html());
  this.renderResponse = _.template($(responseTemplateSelector).html());
  this.renderError = _.template($(errorTemplateSelector).html());
  this.renderSys = _.template($(sysTemplateSelector).html());
}

MessageDisplay.prototype.addRequest = function(request) {
  this.elem.append(this.renderRequest({message: request}));
};

MessageDisplay.prototype.addResponse = function(response) {
  this.elem.append(this.renderResponse({message: response}));
};

MessageDisplay.prototype.addError = function(error) {
  this.elem.append(this.renderError({message: error}));
};

MessageDisplay.prototype.addSys = function(sys) {
  this.elem.append(this.renderSys({message: sys}));
};

var Skeam = function(config) {
  this.inputHandler = new InputHandler(config.inputSelector);
  this.messageDisplay = new MessageDisplay(config.outputSelector,
                                           config.requestTemplateSelector,
                                           config.responseTemplateSelector,
                                           config.errorTemplateSelector,
                                           config.sysTemplateSelector);
  this.conn = new ConnectionHandler(config.wsPath);
  document.addEventListener("sendMsg", _.bind(this.sendMsg, this), false);
  document.addEventListener("receiveResponse", _.bind(this.receiveResponse, this), false);
  document.addEventListener("wsClosed", _.bind(this.onclose, this), false);
};

Skeam.prototype.sendMsg = function(event) {
  this.messageDisplay.addRequest(event.message);
  this.conn.sendMsg(event.message + "\n");
};

Skeam.prototype.receiveResponse = function(event) {
  var message = JSON.parse(event.message);
  if (message.is_error) {
    this.messageDisplay.addError(message.message);
  } else {
    this.messageDisplay.addResponse(message.message);
  }
};

Skeam.prototype.onclose = function() {
  this.messageDisplay.addSys("connection closed");
};
