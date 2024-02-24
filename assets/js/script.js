/* use the sortable library to achieve sorting of divs */
; function sortableRefresh(content) {
    var sortables = content.querySelectorAll(".sortable");
    for (var i = 0; i < sortables.length; i++) {
        var sortable = sortables[i];
        var sortableInstance = new Sortable(sortable, {
            animation: 150,
            onMove: function (evt) {
                return true;
            },
            onEnd: function (evt) {
                document.getElementById('save_list_sort_order').classList.remove('d-none');
            }
        });
    }
}
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
/* resize */
; (function resize() {
    function viewPort(x, y) {
        document.cookie = `viewport=${x}:${y}`;
    }
    window.addEventListener("resize", (event) => {
        console.log(`${window.innerWidth}/${window.innerHeight}`);
        viewPort(window.innerWidth, window.innerHeight);
    });
    window.addEventListener("load", (event) => {
        console.log(`${window.innerWidth}/${window.innerHeight}`);
        viewPort(window.innerWidth, window.innerHeight);
    });
})();
/* enable popover */
; (function initPopover() {
    // const popoverTriggerList = document.querySelectorAll('[data-bs-toggle="popover"]');
    // const popoverList = [...popoverTriggerList].map(popoverTriggerEl => new bootstrap.Popover(popoverTriggerEl));
})();
