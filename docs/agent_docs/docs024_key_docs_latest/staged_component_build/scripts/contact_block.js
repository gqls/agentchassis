(function() {
  var form = document.getElementById('cb-contact-form');
  var submitBtn = document.getElementById('cb-submit-btn');
  var statusEl = document.getElementById('cb-status');

  if (!form) return;

  // WHY THIS FILE LOOKS THE WAY IT DOES (bugs_open/228).
  // Until 2026-08-09 this component validated properly, then ran a 1200ms
  // setTimeout and told the visitor "Your message has been sent" — with no
  // form action and no fetch/XHR anywhere in it. Nothing was ever sent. The
  // rule this rewrite is built on: the success message must be DOWNSTREAM OF A
  // DESTINATION ACCEPTING THE MESSAGE, never downstream of a timer.
  //
  // The destination is the form's own `action`, which the platform fills in
  // per-site: the template carries `action="{{.form_action}}"`, and
  // sanitiseFormAction (component_library.go) replaces a non-delivering value
  // with `mailto:<sites.email>?subject=<domain> enquiry`. So this file never
  // invents an address and never hard-codes one.
  //
  // Three destination shapes, three DIFFERENT and honest outcomes:
  //   http(s):  POST it and report what the server actually said. This is the
  //             shape a real receipt endpoint will use; the component is
  //             already correct for it, so switching is a config change.
  //   mailto:   hand off to the visitor's mail client and say exactly that —
  //             "opening your email app" — because we do not know, and cannot
  //             know, whether they then press send.
  //   anything else (empty, "#", unsanitised): say the form has no destination
  //             and ask them to use the address shown above. NEVER a success.
  //
  // A mailto: is built here, in JS, with explicit subject= and body= params
  // rather than by submitting the form to it. Measured at Chromium 2026-08-09
  // (probe_mailto_form_encoding.go): a GET form REPLACES the action's query, so
  // the ?subject= is destroyed and each field becomes a mail header; a POST
  // form hands the text to a body that a mailto URL cannot carry. Building the
  // URL removes the browser-dependence entirely.

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

  function setBusy(busy, busyLabel) {
    var btnText = submitBtn.querySelector('.cb-btn-text');
    if (busy) {
      submitBtn.disabled = true;
      if (btnText) {
        submitBtn.setAttribute('data-original-label', btnText.textContent);
        btnText.textContent = busyLabel;
      }
      return;
    }
    submitBtn.disabled = false;
    if (btnText && submitBtn.getAttribute('data-original-label') !== null) {
      btnText.textContent = submitBtn.getAttribute('data-original-label');
    }
  }

  form.addEventListener('submit', function(e) {
    e.preventDefault();
    clearStatus();

    var firstName = form.querySelector('#cb-first-name').value.trim();
    var lastName = form.querySelector('#cb-last-name').value.trim();
    var email = form.querySelector('#cb-email').value.trim();
    var subject = form.querySelector('#cb-subject').value;
    var message = form.querySelector('#cb-message').value.trim();

    if (!firstName || !lastName) {
      setStatus('error', 'Please enter your full name.');
      return;
    }
    if (!validateEmail(email)) {
      setStatus('error', 'Please enter a valid email address.');
      return;
    }
    if (!subject) {
      setStatus('error', 'Please select a subject.');
      return;
    }
    if (message.length < 10) {
      setStatus('error', 'Please enter a message of at least 10 characters.');
      return;
    }

    var action = (form.getAttribute('action') || '').trim();

    // No destination. This is the state bugs_open/228 was filed for, and the
    // ONLY honest thing to do is refuse and point at the address on the page.
    if (!action || action === '#' || action.charAt(0) === '#') {
      setStatus('error',
        'This form has no destination configured, so your message was not sent. ' +
        'Please email us using the address shown above — and sorry, that is our fault, not yours.');
      return;
    }

    // ---- real endpoint: the outcome is whatever the server says ----
    if (/^https?:\/\//i.test(action)) {
      setBusy(true, 'Sending…');
      var payload = new FormData(form);
      fetch(action, { method: 'POST', body: payload })
        .then(function(r) {
          setBusy(false);
          if (!r.ok) {
            setStatus('error',
              'We could not send your message (error ' + r.status + '). ' +
              'Your text is still here — please try again, or email us using the address above.');
            return;
          }
          // ONLY here, after a destination accepted it, is "sent" true.
          setStatus('success', 'Your message has been sent. We\'ll be in touch shortly.');
          form.reset();
        })
        .catch(function() {
          setBusy(false);
          setStatus('error',
            'We could not reach the server, so your message was not sent. ' +
            'Your text is still here — please try again, or email us using the address above.');
        });
      return;
    }

    // ---- mailto: hand off to the visitor's own mail client ----
    if (/^mailto:/i.test(action)) {
      var addr = action.slice('mailto:'.length).split('?')[0];
      var existing = action.indexOf('?') >= 0 ? action.slice(action.indexOf('?') + 1) : '';
      var subjectText = '';
      // Honour a subject the platform already put in the action, else build one.
      var m = /(?:^|&)subject=([^&]*)/.exec(existing);
      if (m) {
        subjectText = decodeURIComponent(m[1].replace(/\+/g, ' '));
      }
      if (!subjectText) { subjectText = 'Website enquiry'; }
      subjectText = subjectText + ' — ' + subject;

      var body =
        'Name: ' + firstName + ' ' + lastName + '\n' +
        'Email: ' + email + '\n' +
        'Subject: ' + subject + '\n\n' +
        message + '\n';

      var url = 'mailto:' + addr +
        '?subject=' + encodeURIComponent(subjectText) +
        '&body=' + encodeURIComponent(body);

      // Deliberately NOT form.reset(): we do not know that they sent it, and
      // wiping their text after handing it to a mail client that may not have
      // opened is how a message gets lost with nobody able to retry.
      setStatus('success',
        'Opening your email app with this message ready to send to ' + addr + '. ' +
        'If nothing opens, please email that address directly — your text is still here.');
      window.location.href = url;
      return;
    }

    // A relative or otherwise unrecognised action. Submit it natively rather
    // than guess: the browser knows what to do with a same-origin path, and a
    // navigation is visible to the visitor either way.
    setBusy(true, 'Sending…');
    form.submit();
  });

  var inputs = form.querySelectorAll('input, select, textarea');
  inputs.forEach(function(input) {
    input.addEventListener('input', function() {
      clearStatus();
    });
  });
})();
