  var GoogleAuth;
  var SCOPE = 'openid';

  function handleClientLoad() {
    // Load the API's client and auth2 modules.
    // Call the initClient function after the modules load.
    gapi.load('client:auth2', initClient);
  }

  function initClient() {
     // Initialize the gapi.client object, which app uses to make API requests.
     // Get API key and client ID from API Console.
     // 'scope' field specifies space-delimited list of access scopes.
     gapi.client.init({
        'clientId': '502307674846-9hk8u2iaggriv00op22gk2iqoffmje9d.apps.googleusercontent.com',
	'scope': SCOPE,
	'ux_mode': 'redirect',
	'redirect_uri': 'http://localhost:8080/app/index.html' 
     }).then(function () {

      GoogleAuth = gapi.auth2.getAuthInstance();

      // Listen for sign-in state changes.
      GoogleAuth.isSignedIn.listen(updateSigninStatus);

      // Handle initial sign-in state. (Determine if user is already signed in.)
      var user = GoogleAuth.currentUser.get();
      setSigninStatus();

      // Call handleAuthClick function when user clicks on
      //      "Sign In/Authorize" button.
      $('#nav-signin').click(function() {
        handleAuthClick();
      });
      $('#nav-signout').click(function() {
        revokeAccess();
      });
      $('#nav-statistics').click(function() {
        showStatistics();
      });
    });
  }

  function handleAuthClick() {
    if (GoogleAuth.isSignedIn.get()) {
      // User is authorized and has clicked "Sign out" button.
      GoogleAuth.signOut();
    } else {
      // User is not signed in. Start Google auth flow.
      GoogleAuth.signIn();
    }
  }

  function revokeAccess() {
    GoogleAuth.disconnect();
  }

  function setSigninStatus(isSignedIn) {
	  alert("updating signin status")
    var user = GoogleAuth.currentUser.get();
    var isAuthorized = user.hasGrantedScopes(SCOPE);
    if (isAuthorized) {
      $('#nav-signin').addClass("d-none")
      $('#nav-signout').removeClass("d-none");
      $('#nav-statistics').removeClass('d-none');
    } else {
      $('#nav-signin').removeClass('d-none');
      $('#nav-signout').addClass('d-none');
      $('#nav-statistics').addClass('d-none');
    }
  }

  function updateSigninStatus(isSignedIn) {
    setSigninStatus();
  }

  function showStatistics() {
	  var user = GoogleAuth.currentUser.get();
	  var id_token = user.getAuthResponse().id_token;
	  var xhr = new XMLHttpRequest();
	  xhr.open('GET', '/user/v1/dailystats/id/'+id_token);
	  xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	  xhr.onload = function() {
				   alert(xhr.responseText);
			 };
	  xhr.send();
  }

