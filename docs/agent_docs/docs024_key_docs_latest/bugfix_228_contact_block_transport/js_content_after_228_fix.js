(function() {
  var form = document.getElementById('cb-contact-form');
  var submitBtn = document.getElementById('cb-submit-btn');
  var statusEl = document.getElementById('cb-status');

  if (!form) return;

  function setStatus(type, message) {
    statusEl.className = 'cb-status-message ' + (type === 'success' ? 'cb-success' : 'cb-error');
    statusEl.textContent = message;
  }

  function clearStatus() {
    statusEl.className = 'cb-status-message';
    statusEl.textContent = '';
  }

  function validateEmail(email) {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
  }

  form.addEventListener('submit', function(e) {
    clearStatus();

    var firstName = form.querySelector('#cb-first-name').value.trim();
    var lastName = form.querySelector('#cb-last-name').value.trim();
    var email = form.querySelector('#cb-email').value.trim();
    var subject = form.querySelector('#cb-subject').value;
    var message = form.querySelector('#cb-message').value.trim();

    if (!firstName || !lastName) {
      e.preventDefault();
      setStatus('error', 'Please enter your full name.');
      return;
    }
    if (!validateEmail(email)) {
      e.preventDefault();
      setStatus('error', 'Please enter a valid email address.');
      return;
    }
    if (!subject) {
      e.preventDefault();
      setStatus('error', 'Please select a subject.');
      return;
    }
    if (message.length < 10) {
      e.preventDefault();
      setStatus('error', 'Please enter a message of at least 10 characters.');
      return;
    }

    // Valid: let the browser proceed with its native submission to the
    // form's real action (a mailto: built server-side from the site's
    // configured address). Nothing here can confirm delivery, so the status
    // names the mechanism rather than claiming an outcome, and the form is
    // deliberately NOT reset -- if the visitor's mail client fails to open,
    // their typed text is still there.
    setStatus('success', 'Opening your email client to send this message…');
  });

  var inputs = form.querySelectorAll('input, select, textarea');
  inputs.forEach(function(input) {
    input.addEventListener('input', function() {
      clearStatus();
    });
  });
})();
