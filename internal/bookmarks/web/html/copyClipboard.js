; (function clipboardLogic() {
	const copyClipboards = document.querySelectorAll(".copy-clipboard-btn");
	if (copyClipboards) {
		for (let i = 0; i < copyClipboards.length; i++) {
			const copyClipboard = copyClipboards[i];
			copyClipboard.addEventListener("click", (event) => {
				event.preventDefault();
				try {
					var url = copyClipboard.getAttribute("data-clipboard-text");
					if (url) {
						navigator.clipboard.writeText(url);
						//alert(`Copied '${url}' to clipboard!`)
						var body = {
							"notification_message": `Copied '${url}' to clipboard!`
						};
						htmx.ajax('POST', '/bm/toast', { swap: 'none', values: body });
					} else {
						console.log("could not get a URL to copy");
					}
				} catch (error) {
					alert("Could not copy URL to clipboard!");
					console.error(error.message);
				}
			});
		}
	}
})();

