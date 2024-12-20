if (document.querySelector('#type_Bookmark')) {
    document.querySelector('#type_Bookmark').addEventListener('change', (event) => {
        if (event.target.value === 'Node') {
            document.querySelector('#url_section').classList.remove('d-none');
        } else {
            document.querySelector('#url_section').classList.add('d-none');
        }
    });
}

if (document.querySelector('#type_Folder')) {
    document.querySelector('#type_Folder').addEventListener('change', (event) => {
        if (event.target.value === 'Folder') {
            document.querySelector('#url_section').classList.add('d-none');
        } else {
            document.querySelector('#url_section').classList.remove('d-none');
        }
    });
}

if (document.querySelector('#bookmark_Custom_Favicon')) {
    document.querySelector('#bookmark_Custom_Favicon').addEventListener('change', (event) => {
        if (event.currentTarget.checked) {
            document.querySelector('#custom_favicon_section').classList.remove('d-none');
        } else {
            document.querySelector('#custom_favicon_section').classList.add('d-none');
        }
    });
}

if (document.querySelector('#bookmark_Invert')) {
    document.querySelector('#bookmark_Invert').addEventListener('change', (event) => {
        if (event.currentTarget.checked) {
            document.querySelector('#bookmark_favicon_display').classList.add('invert');
        } else {
            document.querySelector('#bookmark_favicon_display').classList.remove('invert');
        }
    });
}

if (document.querySelector('.bookmark_edit_form')) {
    document.querySelector('.bookmark_edit_form').addEventListener('paste', e => {
        if (!e.clipboardData.items || e.clipboardData.items.length == 0) {
            showInfoText(`Nothing to paste from clipboard!`);
            return;
        }
        try {
            // get the first item of the clipboard
            var item = e.clipboardData.items[0];
            if (item.type.indexOf("image") === 0 || item.type.indexOf("svg") === 0) {
                let fileInput = document.querySelector('#customFaviconUpload');
                let dataTransfer = new DataTransfer();
                let blob = item.getAsFile();
                let uuid = window.crypto.randomUUID()
                dataTransfer.items.add(blob);
                fileInput.files = dataTransfer.files;

                console.log('files for upload: ' + fileInput.files.length);
                showInfoText(`Pasted file '${blob.name}' from clipboard!`);
            } else {
                showInfoText(`No image in clipboard!`);
            }
        } catch (e) {
            console.log("could not set clipboard image!");
            console.log(e);
        }

    })
}

function showInfoText(text) {
    if (document.querySelector('#info_section')) {
        if (document.querySelector('#info_section_text')) {
            document.querySelector('#info_section_text').textContent = text;
            document.querySelector('#info_section').classList.remove('d-none');

            setTimeout(() => {
                document.querySelector('#info_section_text').textContent = '';
                document.querySelector('#info_section').classList.add('d-none')
            }, 2000);
        }
    }
}
