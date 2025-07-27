; (function __ageLogic() {

	const agePerformAction = document.querySelector('#age_perform_action');
	agePerformAction.addEventListener("click", (event) => {
		event.preventDefault();
		try {
			document.getElementById('request_indicator').classList.remove('htmx-indicator');
			let passphrase = document.querySelector('#age_passphrase').value;
			let input = document.querySelector('#age_input').value;
			let output = document.querySelector('#age_output').value;

			if (!passphrase || passphrase === '') {
				var body = {
					"notification_message": `Missing Passphrase!`
				};
				htmx.ajax('PUT', '/age/toast', { swap: 'none', values: body });

				return;
			}

			// when an input is available, use the input for encryption
			if (input && input !== '') {
				(async () => {
					const e = new age.Encrypter();
					e.setPassphrase(passphrase);
					const ciphertext = await e.encrypt(input);
					const armored = age.armor.encode(ciphertext)
					document.querySelector('#age_output').value = armored;

					var body = {
						"notification_message": `Encrypted!`,
					};
					htmx.ajax('PUT', '/age/toast', { swap: 'none', values: body });

					document.getElementById('request_indicator').classList.add('htmx-indicator');
				})()
			}

			// when an input is available, use the input for encryption
			if (output && output !== '') {
				(async () => {
					const d = new age.Decrypter();
					d.addPassphrase(passphrase);
					// the input will be armored
					const dearmored = age.armor.decode(output)
					const decrypted = await d.decrypt(dearmored, "text")
					document.querySelector('#age_input').value = decrypted;

					var body = {
						"notification_message": `Decrypted!`,
					};
					htmx.ajax('PUT', '/age/toast', { swap: 'none', values: body });

					document.getElementById('request_indicator').classList.add('htmx-indicator');
				})()
			}

		} catch (error) {
			document.querySelector('#request_indicator').classList.add('htmx-indicator')
			var body = {
				"notification_message": `Error: ` + error,
				"notification_type": "error",
			};
			htmx.ajax('PUT', '/age/toast', { swap: 'none', values: body });
			console.error(error.message);
		}
	});

})();
