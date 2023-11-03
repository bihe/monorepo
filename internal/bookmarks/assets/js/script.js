/* use the sortable library to achieve sorting of divs */
; function sortableRefresh(content) {
    var sortables = content.querySelectorAll(".sortable");
    for (var i = 0; i < sortables.length; i++) {
        var sortable = sortables[i];
        var sortableInstance = new Sortable(sortable, {
            animation: 150,
            ghostClass: 'blue-background-class',

            onMove: function (evt) {
                //return evt.related.className.indexOf('htmx-indicator') === -1;
                return true;
            },

            // Disable sorting on the `end` event
            onEnd: function (evt) {
                //this.option("disabled", true);
                // enable the button to save the new sort-order
                document.getElementById('save_list_sort_order').classList.remove('d-none');
            }
        });
    }
}

(function sortableStartup() {
    htmx.onLoad(function (content) {
        sortableRefresh(content);
    });
})();

/* logic for toast-messages */
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
