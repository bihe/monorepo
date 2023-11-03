function tryConnectToReload(address) {
    const wsConn = 'ws://%s/reload';
    // This is a statically defined port on which the app is hosting the reload service.
    let conn = new WebSocket(wsConn);

    conn.onclose = function(evt) {
      // The reload endpoint hasn't been started yet, we are retrying in 2 seconds.
      setTimeout(() => tryConnectToReload(), 2000);
    };

    conn.onmessage = function(evt) {
      console.log('Refresh received!');
      location.reload()
    };
    console.log('connected to reload server: '+wsConn);
  }

  try {
    if (window["WebSocket"]) {
      tryConnectToReload();
    } else {
      console.log("Your browser does not support WebSocket, cannot connect to the reload service.");
    }
  } catch (ex) {
    console.log('Exception during connecting to reload:', ex);
  }
