function signOut() {
	firebase.auth().signOut();
  	window.location.replace("/logout");
}

function signIn() {
	var email = emailElement.value;
	var password = passwordElement.value;
    if (email.length < 4) {
      alert('Please enter an email address.');
      return;
    }
    if (password.length < 4) {
      alert('Please enter a password.');
      return;
    }
    firebase.auth().signInWithEmailAndPassword(email, password)
	.then(authUser => {
      if(authUser.user.emailVerified){ //This will return true or false
       console.log('email is verified')
      }else{
		alert('email not verified');
		signOut();
      }
	}).catch(function(error) {
      // Handle Errors here.
      var errorCode = error.code;
      var errorMessage = error.message;
      if (errorCode === 'auth/wrong-password') {
        alert('Wrong password.');
      } else {
        alert(errorMessage);
      }
      console.log(error);
    });
}

// Returns the signed-in user's display name.
function getUserName() {
	user = firebase.auth().currentUser
	if (user != null) {	
		return user.displayName;
	}
	return '';
}

// Initiate firebase auth.
function initFirebaseAuth() {
	firebase.auth().onAuthStateChanged(authStateObserver);
}

// Triggers when the auth state change for instance when the user signs-in or signs-out.
function authStateObserver(user) {
  if (user) { // User is signed in!
  	if (!user.emailVerified) {
		signOut();
		return;
	}
	var userName = getUserName();
	// Set the user's profile name.
	userNameElement.textContent = userName;
	// Show user's profile and sign-out button.
	userNameElement.removeAttribute('hidden');
	profilesLinkElement.removeAttribute('hidden');
	signOutButtonElement.removeAttribute('hidden');
	// Hide elements.
	signInButtonElement.setAttribute('hidden', 'true');
	signUpButtonElement.setAttribute('hidden', 'true');
	emailElement.setAttribute('hidden', 'true');
	passwordElement.setAttribute('hidden', 'true');
	emailSignupElement.setAttribute('hidden', 'true');
	passwordSignupElement.setAttribute('hidden', 'true');
	nameSignupElement.setAttribute('hidden', 'true');
	birthdayDivSignupElement.setAttribute('hidden', 'true');
	// Get the user's ID token as it is needed to exchange for a session cookie.
	return user.getIdToken().then(idToken => {
	  // Session login endpoint is queried and the session cookie is set.
	  // CSRF protection should be taken into account.
	  // ...
	  const csrfToken = getCookie('csrfToken')
	  return postIdTokenToSessionLogin('/login', idToken, csrfToken);
	});
  } else { // User is signed out!
    // Hide user's profile and sign-out button.
    userNameElement.setAttribute('hidden', 'true');
    profilesLinkElement.setAttribute('hidden', 'true');
    signOutButtonElement.setAttribute('hidden', 'true');
    // Show elements.
    signInButtonElement.removeAttribute('hidden');
    signUpButtonElement.removeAttribute('hidden');
	emailElement.removeAttribute('hidden');
	passwordElement.removeAttribute('hidden');
	emailSignupElement.removeAttribute('hidden');
	passwordSignupElement.removeAttribute('hidden');
	nameSignupElement.removeAttribute('hidden');
	birthdayDivSignupElement.removeAttribute('hidden');
  }
}

function getCookie(name) {
	var v = document.cookie.match('(^|;) ?' + name + '=([^;]*)(;|$)');
	return v ? v[2] : null;
}

function postIdTokenToSessionLogin(url, idToken, csrfToken) {
  // POST to session login endpoint.
  return $.ajax({
	type:'POST',
	url: url,
	data: {idToken: idToken, csrfToken: csrfToken},
	contentType: 'application/x-www-form-urlencoded'
  });
}

// Shortcuts to DOM Elements.
var userNameElement = document.getElementById('user-name');
var profilesLinkElement = document.getElementById('profiles-link');
var emailElement = document.getElementById('email');
var passwordElement = document.getElementById('password');

var emailSignupElement = document.getElementById('email-signup');
var passwordSignupElement = document.getElementById('password-signup');
var nameSignupElement = document.getElementById('name-signup');
var birthdaySignupElement = document.getElementById('birthday-signup');
var birthdayDivSignupElement = document.getElementById('birthday-div');

var signInButtonElement = document.getElementById('sign-in');
var signOutButtonElement = document.getElementById('sign-out');
var signUpButtonElement = document.getElementById('sign-up');

signOutButtonElement.addEventListener('click', signOut);
signInButtonElement.addEventListener('click', signIn);

initFirebaseAuth();
