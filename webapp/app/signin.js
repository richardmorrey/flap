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
        waitCursorOn();
        showCharts();
      });  
      $('#nav-planning').click(function() {
	waitCursorOn();
	showPlanning();
      });
      $('#nav-history').click(function() {
       waitCursorOn();
       showHistory();
      });
      $('#nav-account').click(function() {
        navbarActive('account')
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

var gPages=['signin','signout','statistics','planning','history','account']
var gPageTitles=['Welcome','Welcome','Statistics','Trip Planning','Flight History','Distance Account']

  function setSigninStatus(isSignedIn) {
    var user = GoogleAuth.currentUser.get();
    var isAuthorized = user.hasGrantedScopes(SCOPE);
    for (i in gPages) {
	if ((isAuthorized && (gPages[i]!='signin')) || (!isAuthorized && gPages[i]=='signin'))
	{
      		$('#nav-'+gPages[i]).removeClass("d-none")
	} else {
      		$('#nav-'+gPages[i]).addClass("d-none")
    	}
    }
    if (isAuthorized) {
	showCharts()
	$("#useremail").text(user.getBasicProfile().getEmail())
	
    } else {
	$("#useremail").text("")    
	navbarActive("signout")
    }
  }

  function updateSigninStatus(isSignedIn) {
    setSigninStatus();
  }

  var gPage
  function navbarActive(opt) {
	for (i in gPages) {
		if (opt == gPages[i]) {
			$('#nav-'+gPages[i]).css('fill','')
			$('#pg_'+gPages[i]).addClass("flapfadinelement")
			$('#pg_'+gPages[i]).removeClass("d-none")
			$("#pagetitle").text(gPageTitles[i])    
		} else {
			$('#nav-'+gPages[i]).css('fill','white')
			$('#pg_'+gPages[i]).addClass("d-none")
		}
	}
	gPage=opt
	waitCursorOff()
  }

  function waitCursorOn() {
	document.body.classList.add('inheritCursors');
	document.body.style.cursor = 'progress';
  }

  function waitCursorOff() {
	document.body.classList.remove('inheritCursors');
        document.body.style.cursor = 'unset';
  }


  var gBotBand=5
  var gBotNumber=1
  $('#changeModal').on('hidden.bs.modal', function (e) {

	  gBotNumber = $('#tNum').val()
	  gBotBand = $("#tBand").prop('selectedIndex')
	  historyInit=false
	  planningInit=false
	  $('#nav-'+gPage).trigger("click");
  })

