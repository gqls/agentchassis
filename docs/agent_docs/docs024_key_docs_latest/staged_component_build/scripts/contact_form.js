(function() {
  var form = document.getElementById('cf-contact-form');
  var statusEl = document.getElementById('cf-status');
  if (!form || !statusEl) return;

  // Sibling of contact-block.js and DELIBERATELY the same logic — see that file
  // for the full reasoning (bugs_open/228). The short version:
  //
  //   the destination is the form's own `action`, filled per-site by
  //   sanitiseFormAction; a success message is only printed when a destination
  //   ACCEPTED the message; a mailto: is built here with explicit subject= and
  //   body= rather than by submitting the form to it.
  //
  // Why this component needed changing even though it was never dishonest:
  // it submitted natively to `action="mailto:…" method="POST"`. Measured at
  // Chromium 2026-08-09 (probe_mailto_form_encoding.go), that hands the typed
  // text to a request BODY, and a mailto: URL has no body — so whether the
  // message survives is the browser's business, not ours. With GET it is worse:
  // the form data REPLACES the action's query, destroying the ?subject= the
  // platform put there. Building the URL removes the guesswork.
  //
  // Two files rather than one shared module because components ship their own
  // asset and there is no import path between them. If a third contact
  // component appears, promote this into a snippet rather than copy it again.

  var submitBtn = form.querySelector('.form-submit');

  function setStatus(type, message) {
    statusEl.className = 'contact-form-status ' + (type === 'success' ? 'cf-success' : 'cf-error');
    statusEl.textContent = message;
  }

  function clearStatus() {
    statusEl.className = 'contact-form-status';
    statusEl.textContent = '';
  }

  function validateEmail(email) {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
  }

  form.addEventListener('submit', function(e) {
    e.preventDefault();
    clearStatus();

    var name = form.querySelector('#name').value.trim();
    var email = form.querySelector('#email').value.trim();
    var message = form.querySelector('#message').value.trim();

    if (!name) {
      setStatus('error', 'Please enter your name.');
      return;
    }
    if (!validateEmail(email)) {
      setStatus('error', 'Please enter a valid email address.');
      return;
    }
    if (message.length < 10) {
      setStatus('error', 'Please enter a message of at least 10 characters.');
      return;
    }

    var action = (form.getAttribute('action') || '').trim();

    if (!action || action.charAt(0) === '#') {
      setStatus('error',
        'This form has no destination configured, so your message was not sent. ' +
        'Please contact us directly — and sorry, that is our fault, not yours.');
      return;
    }

    if (/^https?:\/\//i.test(action)) {
      if (submitBtn) { submitBtn.disabled = true; }
      fetch(action, { method: 'POST', body: new FormData(form) })
        .then(function(r) {
          if (submitBtn) { submitBtn.disabled = false; }
          if (!r.ok) {
            setStatus('error',
              'We could not send your message (error ' + r.status + '). ' +
              'Your text is still here — please try again.');
            return;
          }
          setStatus('success', 'Your message has been sent. We\'ll be in touch shortly.');
          form.reset();
        })
        .catch(function() {
          if (submitBtn) { submitBtn.disabled = false; }
          setStatus('error',
            'We could not reach the server, so your message was not sent. ' +
            'Your text is still here — please try again.');
        });
      return;
    }

    if (/^mailto:/i.test(action)) {
      var addr = action.slice('mailto:'.length).split('?')[0];
      var existing = action.indexOf('?') >= 0 ? action.slice(action.indexOf('?') + 1) : '';
      var subjectText = '';
      var m = /(?:^|&)subject=([^&]*)/.exec(existing);
      if (m) { subjectText = decodeURIComponent(m[1].replace(/\+/g, ' ')); }
      if (!subjectText) { subjectText = 'Website enquiry'; }

      var body = 'Name: ' + name + '\nEmail: ' + email + '\n\n' + message + '\n';
      var url = 'mailto:' + addr +
        '?subject=' + encodeURIComponent(subjectText) +
        '&body=' + encodeURIComponent(body);

      setStatus('success',
        'Opening your email app with this message ready to send to ' + addr + '. ' +
        'If nothing opens, please email that address directly — your text is still here.');
      window.location.href = url;
      return;
    }

    form.submit();
  });

  var inputs = form.querySelectorAll('input, textarea');
  inputs.forEach(function(input) {
    input.addEventListener('input', clearStatus);
  });
})();
