try {
    document.querySelector('#btn_toggle_sorting').addEventListener('click', (event) => {
        if (event.target.classList.contains('active')) {
            console.log('Activate sorting');
            sortableRefresh(document);
        } else {
            console.log('Disable sorting - refresh the list');
            document.querySelector('#save_list_sort_order').classList.add('d-none');
            htmx.trigger('#btn_toggle_sorting', 'refreshBookmarkList');
        }
    });
    document.querySelector('#btn_save_sorting').addEventListener('click', (event) => {
        htmx.trigger('#btn_save_sorting', 'sortBookmarkList');
        document.querySelector('#btn_toggle_sorting').classList.remove('active');
        document.querySelector('#save_list_sort_order').classList.add('d-none');
    });
} catch (error) {
    console.error(error);
}
