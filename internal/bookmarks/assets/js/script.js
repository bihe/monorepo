; (function toastLogic() {
    document.body.addEventListener("toastMessage", function (evt) {
        const toastType = evt.detail.type;
        document.getElementById('toast_message_title-' + toastType).textContent = evt.detail.title;
        document.getElementById('toast_messsage_text-' + toastType).textContent = evt.detail.text;
        document.getElementById('toastMessage-' + toastType).classList.add('show');

        if (toastType === 'success') {
            setTimeout(() => {
                // time to say goodbye
                document.getElementById('toastMessage-' + toastType).classList.remove('show');
            }, 2500);
        }
    });
})();
